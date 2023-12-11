package signature

import (
	"fmt"
	"os"
	"os/exec"
)

// LocalCryptoPro реализация [signature.Provider] с использованием
// утилиты csptest из локально установленного пакета КриптоПро CSP для рабочих станций версии 5 и выше.
// Для подписания используется алгоритм ГОСТ Р 34.10-2012 (256 бит).
//
// ВАЖНО: используйте эту реализацию только для отладки взаимодействия с ЕСИА,
// тк КриптоПро CSP 5 для рабочих станций не может использоваться в качестве серверного решения.
type LocalCryptoPro struct {
	cspTestPath  string
	cspContainer string
	certHash     string
	cmd          cmdInterface
}

// NewLocalCryptoPro - конструктор LocalCryptoPro.
//
// # cspTestPath
//
// Полный путь к утилите csptest из пакета КриптоПро CSP:
//   - Mac: "/opt/cprocsp/bin/csptest"
//   - Win: "C:\Program Files\Crypto Pro\CSP\сsptest.exe"
//
// # cspContainer
//
// Имя контейнера сертификата.
// Сертификат ИС (6 файлов .key), используемый для подписи запросов к ЕСИА,
// должен быть записан на съемный носитель (флешку).
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
// Подключите съемный носитель с сертификатом
// и запустите утилиту cpverify (cpverify.exe для Windows) из пакета КриптоПро CSP:
//
//	cpverify -mk <path/to/cert.cer> -alg GR3411_2012_256
//
// Команда выведет хеш сертификата:
//
//	1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF0
func NewLocalCryptoPro(cspTestPath, cspContainer, certHash string) *LocalCryptoPro {
	return &LocalCryptoPro{
		cspTestPath:  cspTestPath,
		cspContainer: cspContainer,
		certHash:     certHash,
		cmd:          osExec{},
	}
}

// CertHash возвращает хэш сертификата
func (p *LocalCryptoPro) CertHash() string {
	return p.certHash
}

// Sign - возвращает подпись для данных c использованием алгоритма ГОСТ Р 34.10-2012 (256 бит).
func (p *LocalCryptoPro) Sign(data []byte) ([]byte, error) {
	// создаем временный файл для подписываемых данных
	dataTempFile, err := os.CreateTemp("", "data")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}
	//goland:noinspection ALL
	defer os.Remove(dataTempFile.Name())

	// создаем временный файл для подписи
	signTempFile, err := os.CreateTemp("", "signature")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileCreate, err)
	}
	//goland:noinspection ALL
	defer os.Remove(signTempFile.Name())

	// записываем подписываемые данные во временный файл
	err = os.WriteFile(dataTempFile.Name(), data, 0644)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileWrite, err)
	}

	// вызываем утилиту csptest, которая подписывает данные
	// и записывает подпись во временный файл

	err = p.cmd.Run(p.cspTestPath,
		"-keys",
		"-sign", "GOST12_256",
		"-cont", p.cspContainer,
		"-keytype", "exchange",
		"-in", dataTempFile.Name(),
		"-out", signTempFile.Name(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCPTestExec, err)
	}

	// читаем подпись из временного файла
	signBytes, err := os.ReadFile(signTempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTempFileRead, err)
	}

	// реверсируем байты подписи, тк это в данном случае требуется
	// http://www.gogost.cypherpunks.ru/FAQ.html
	for i, j := 0, len(signBytes)-1; i < j; i, j = i+1, j-1 {
		signBytes[i], signBytes[j] = signBytes[j], signBytes[i]
	}

	return signBytes, nil
}

type cmdInterface interface {
	Run(path string, args ...string) error
}

type osExec struct{}

func (o osExec) Run(path string, args ...string) error {
	return exec.Command(path, args...).Run()
}
