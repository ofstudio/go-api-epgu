// Пример запроса согласия у пользователя и получения маркера доступа ЕСИА
// для работы с API Госуслуг (АПИ ЕПГУ).
//
// # Основные шаги процесса
//  1. Создание ссылки на страницу предоставления прав доступа ЕСИА (/oauth2/v2/ac)
//  2. Переход пользователя по ссылке
//  3. Получение авторизационного кода из параметров обратного вызова на redirect_uri
//  4. Обмен авторизационного кода на маркер доступа (/oauth2/v3/te)
//
// # Требования
//  1. Информационная система должна быть зарегистрирована на
//     Технологическом портале ЕСИА: продуктовом или тестовом (SVCDEV)
//  2. Для ИС должен быть выпущен необходимый сертификат
//  3. Публичная часть сертификата должна быть загружена на Технологический портал ЕСИА
//  4. Выполнены все необходимые шаги регламента подключения ИС к тестовой
//     или продуктовой среде ЕСИА и согласована заявка на доступ ИС к необходимым скоупам
//  5. Доступ к учетной записи пользователя на тестовом (SVCDEV) или продуктовом портале Госуслуг
//  6. Локально установленный пакет КриптоПро CSP (https://www.cryptopro.ru/products/csp)
//  7. Сертификат ИС (6 файлов .key) записанный на съемном носителе (флешке)
//     для работы с КриптоПро CSP
//
// # Адреса Технологического портала ЕСИА
//   - Тестовая среда (SVCDEV): https://esia-portal1.test.gosuslugi.ru/console/tech
//   - Продуктовая среда: https://esia.gosuslugi.ru/console/tech/
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
//
// # Страница предоставленных согласий пользователя на Портале Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-betalk.test.gosuslugi.ru/settings/third-party/agreements/acting
//   - Продуктовая среда: https://lk.gosuslugi.ru/settings/third-party/agreements/acting
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ofstudio/go-api-epgu/esia/aas"
	"github.com/ofstudio/go-api-epgu/esia/signature"
	"github.com/ofstudio/go-api-epgu/utils"
)

// Параметры подключения к ЕСИА.
// Адреса ЕСИА:
//   - Тестовая среда (SVCDEV): https://esia-portal1.test.gosuslugi.ru
//   - Продуктовая среда: https://esia.gosuslugi.ru
//
// ВАЖНО: значения полей вида "<< поле >>" необходимо заполнить актуальными данными вашей ИС.
const (
	mnemonic    = "<< мнемоника ИС потребителя >>" // Мнемоника ИС на портале ЕСИА
	esiaURI     = "<< адрес ЕСИА >>"               // Адрес портала ЕСИА
	redirectURI = "http://localhost:8000/callback" // Адрес redirect_uri на стороне потребителя
)

// Параметры КриптоПро для signature.LocalCryptoPro.
//
// Ссылка на страницу предоставления прав доступа ЕСИА должна содержать параметры
//   - client_secret с электронной подписью ИС потребителя
//   - client_certificate_hash с хешем сертификата
//
// Подробнее см "Методические рекомендации по использованию ЕСИА",
// раздел "Получение авторизационного кода (v2/ac)".
//
// В данном примере для формирования подписи используется реализация signature.LocalCryptoPro,
// которая вызывает утилиту csptest из локально установленного пакета КриптоПро CSP.
//
// ВАЖНО: значения полей вида "<< поле >>" необходимо заполнить актуальными
// данными вашего сертификата.
const (
	// cspTestPath - полный путь к утилите csptest из пакета КриптоПро CSP:
	//	- Mac: "/opt/cprocsp/bin/csptest"
	//	- Win: "C:\Program Files\Crypto Pro\CSP\сsptest.exe"
	cspTestPath = "<< полный путь к утилите csptest >>"

	// cspContainer - имя контейнера сертификата.
	// Сертификат ИС (6 файлов .key), используемый для подписи запросов к ЕСИА,
	// должен быть записан на съемный носитель (флешку).
	//
	// Для получения имени контейнера, подключите съемный носитель с сертификатом
	// и запустите утилиту csptest (csptest.exe для Windows) из пакета КриптоПро CSP:
	//     csptest -keyset
	// Команда выведет имя контейнера:
	//     ...
	//     Container name: "X9X1XYZA9EZZWZ42"
	//     ...
	cspContainer = "<< имя контейнера >>"

	// certHash - хеш сертификата.
	//
	// Подключите съемный носитель с сертификатом
	// и запустите утилиту cpverify (cpverify.exe для Windows) из пакета КриптоПро CSP:
	//		cpverify -mk <path/to/cert.cer> -alg GR3411_2012_256
	// Команда выведет хеш сертификата:
	//		1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF0
	certHash = "<< хеш сертификата >>"
)

// permissions - параметр ссылки на страницу предоставления прав доступа,
// описывающий тип согласия, цели согласия, список запрашиваемых разрешений (scopes)
// и действия с данными.
//
// Подробнее см "Методические рекомендации по интеграции с REST API Цифрового профиля",
// раздел "Структура JSON-объекта параметра permissions".
//
// ВАЖНО: значения полей вида "<< поле >>" необходимо заполнить актуальными данными вашей ИС.
var permissions = aas.Permissions{
	{
		ResponsibleObject: "<< название организации в ЕСИА >>",                      // Ответственный объект (название организации)
		Sysname:           "APIPGU",                                                 // Тип согласия
		Expire:            525600,                                                   // 1 год: макс. срок согласия данного типа
		Actions:           []aas.PermissionAction{{Sysname: "ALL_ACTIONS_TO_DATA"}}, // Действия с данными
		Purposes:          []aas.PermissionPurpose{{Sysname: "APIPGU"}},             // Цели согласий
		Scopes: []aas.PermissionScope{
			// Скоуп, необходимый для работы с API Госуслуг
			{Sysname: "http://lk.gosuslugi.ru/api-order"},
			// Дополнительно, для данного типа согласия можно запросить следующие скоупы
			//{Sysname: "snils"},
			//{Sysname: "id_doc"},
			//{Sysname: "gender"},
			//{Sysname: "fullname"},
			//{Sysname: "birthdate"},
			//{Sysname: "addresses"},
		},
	},
}

func main() {
	// Создаем провайдер электронной подписи запросов
	signer := signature.NewLocalCryptoPro(cspTestPath, cspContainer, certHash)

	// Создаем клиент ЕСИА
	oauthClient := aas.
		NewClient(esiaURI, mnemonic, signer).
		WithDebug(log.Default()) // Опция включает полное логирование запросов и ответов к ЕСИА

	// === ШАГ 1 ===
	// Создание ссылки на страницу предоставления прав доступа (/oauth2/v2/ac)
	uri, err := oauthClient.AuthURI("openid", redirectURI, permissions)
	if err != nil {
		log.Fatal(err)
	}

	// === ШАГ 2 ===
	// Переход пользователя по ссылке (необходимо вручную перейти по ссылке)
	log.Print("Ссылка на страницу предоставления прав доступа ЕСИА: ", uri)

	// === ШАГ 3 ===
	// Получение авторизационного кода из параметров обратного вызова на redirect_uri
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		message := "=== Запрос к redirect_uri ===\n\n" + utils.PrettyQuery(r.URL.Query())

		code, _, err := oauthClient.ParseCallback(r.URL.Query())
		if err != nil {
			log.Print(err)
			http.Error(w, message+"\nError: "+err.Error(), http.StatusBadRequest)
			return
		}

		// === ШАГ 4 ===
		// Обмен авторизационного кода на маркер доступа (/oauth2/v3/te)
		message += "\n=== Обмен авторизационного кода на маркер доступа ===\n\n"
		res, err := oauthClient.TokenExchange(code, "openid", redirectURI)
		if err != nil {
			log.Print(err)
			http.Error(w, message+err.Error(), http.StatusInternalServerError)
			return
		}

		message += utils.PrettyJSON(res)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, message)
		log.Print("Получен маркер доступа: ", res.AccessToken)
	})

	// Запускаем сервер, который будет слушать запросы к redirectURI
	log.Print("Listening " + redirectURI)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
