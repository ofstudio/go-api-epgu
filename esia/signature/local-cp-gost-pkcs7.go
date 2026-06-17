package signature

import (
	"fmt"
	"os"
)

// LocalCryptoProPKCS7 реализация [signature.Provider] с использованием
// утилиты csptest из локально установленного пакета КриптоПро CSP для рабочих станций версии 5 и выше.
// Создается PKCS#7-подпись с алгоритмом ГОСТ Р 34.10-2012 (256 бит).
//
// ВАЖНО: используйте эту реализацию только для отладки взаимодействия с ЕСИА,
// тк КриптоПро CSP 5 для рабочих станций не может использоваться в качестве серверного решения.
type LocalCryptoProPKCS7 struct {
	cspTestPath    string
	cspContainer   string
	certHash       string
	certThumbprint string
	cmd            cmdInterface
}

// NewLocalCryptoProPKCS7 - конструктор LocalCryptoProPKCS7.
//
// # cspTestPath
//
// Полный путь к утилите csptest из пакета КриптоПро CSP:
//   - Mac: "/opt/cprocsp/bin/csptest"
//   - Win: "C:\Program Files\Crypto Pro\CSP\csptest.exe"
//
// # cspContainer
//
// Имя контейнера сертификата.
// Сертификат ИС (6 файлов .key), используемый для подписи запросов к ЕСИА,
// должен быть записан на съемный носитель (флешку).
// При PKCS#7-подписании сертификат выбирается по certThumbprint, но имя контейнера
// сохраняется в параметрах для единообразия с [LocalCryptoPro].
// Для получения имени контейнера, подключите съемный носитель с сертификатом
// и запустите утилиту csptest (csptest.exe для Windows) из пакета КриптоПро CSP:
//
//	csptest -keyset
//
// Команда выведет имя контейнера:
//
//	...
//	Container name: "X9X1XYZA9EZZWZ42"
//	...
//
// # certHash
//
// Хеш сертификата.
// Сертификат может храниться и загружаться в карточку ИС на ЕСИА в PEM- или
// DER-представлении. Хеш client_certificate_hash должен быть вычислен от
// DER-представления сертификата. Если вычислить хеш от PEM-файла целиком,
// ЕСИА вернет ошибку ESIA-007053: OAuthErrorEnum.clientSecretWrong.
//
// Windows:
//
//	openssl x509 -in "C:\path\to\cert.cer" -outform der | "C:\Program Files\Crypto Pro\CSP\cpverify.exe" -mk -stdin -alg GR3411_2012_256
//
// Linux:
//
//	openssl x509 -in "/path/to/cert.cer" -outform der | /opt/cprocsp/bin/amd64/cpverify -mk -stdin -alg GR3411_2012_256
//
// macOS:
//
//	openssl x509 -in "/path/to/cert.cer" -outform der | /opt/cprocsp/bin/cpverify -mk -stdin -alg GR3411_2012_256
//
// Команда выведет хеш сертификата:
//
//	1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF0
//
// # certThumbprint
//
// Отпечаток сертификата SHA-1, по которому csptest выбирает сертификат для
// PKCS#7-подписи. Для получения отпечатка установите сертификат в хранилище
// пользователя и выполните:
//
//	certmgr -list -store uMy
//
// Команда выведет SHA1 Thumbprint сертификата:
//
//	1234567890abcdef1234567890abcdef12345678
func NewLocalCryptoProPKCS7(cspTestPath, cspContainer, certHash, certThumbprint string) *LocalCryptoProPKCS7 {
	return &LocalCryptoProPKCS7{
		cspTestPath:    cspTestPath,
		cspContainer:   cspContainer,
		certHash:       certHash,
		certThumbprint: certThumbprint,
		cmd:            osExec{},
	}
}

// CertHash возвращает хэш сертификата
func (p *LocalCryptoProPKCS7) CertHash() string {
	return p.certHash
}

// Sign - возвращает отделенную PKCS#7-подпись для данных c использованием
// алгоритма ГОСТ Р 34.10-2012 (256 бит).
func (p *LocalCryptoProPKCS7) Sign(data []byte) ([]byte, error) {
	// создаем временный файл для подписываемых данных
	dataTempFile, err := os.CreateTemp("", "data")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}
	//goland:noinspection ALL
	defer os.Remove(dataTempFile.Name())
	if err = dataTempFile.Close(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}

	// создаем временный файл для подписи
	signTempFile, err := os.CreateTemp("", "signature")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}
	//goland:noinspection ALL
	defer os.Remove(signTempFile.Name())
	if err = signTempFile.Close(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}

	// записываем подписываемые данные во временный файл
	err = os.WriteFile(dataTempFile.Name(), data, 0644)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileWrite, err)
	}

	// вызываем утилиту csptest, которая подписывает данные
	// и записывает подпись во временный файл

	out, err := p.cmd.Run(p.cspTestPath,
		"-sfsign", "-sign", "-detached", "-add",
		"-alg", "GOST12_256",
		"-in", dataTempFile.Name(),
		"-out", signTempFile.Name(),
		"-my", p.certThumbprint,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %s", ErrCPTestExec, err, string(out))
	}

	signBytes, err := os.ReadFile(signTempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileRead, err)
	}

	return signBytes, nil
}
