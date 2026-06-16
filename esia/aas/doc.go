// OAuth2-клиент для запроса согласия и маркера доступа ЕСИА
// для получателей услуг ЕПГУ — физических лиц.
//
// # Методы
//
//   - [Client.AuthURI] — формирует ссылку на страницу ЕСИА и возвращает state для проверки callback-запроса
//   - [Client.AuthURIPKCS7] — формирует ссылку на страницу ЕСИА с PKCS#7-подписью client_secret и возвращает state для проверки callback-запроса
//   - [Client.ParseCallback] — возвращает код авторизации из callback-запроса к redirect_uri и может проверять ожидаемый state
//   - [Client.TokenExchange] — обменивает код авторизации на маркер доступа (токен)
//   - [Client.TokenExchangePKCS7] — обменивает код авторизации на маркер доступа (токен) с PKCS#7-подписью client_secret
//   - [Client.TokenUpdate] — обновляет маркер доступа по идентификатору пользователя (OID)
//   - [Client.TokenUpdatePKCS7] — обновляет маркер доступа по идентификатору пользователя (OID) с PKCS#7-подписью client_secret
//
// Основные методы без суффикса PKCS7 используют raw-подпись client_secret:
// в параметр передается Base64URL-кодированная подпись.
//
// Методы с суффиксом PKCS7 используют эндпоинты ЕСИА /aas/oauth2/ac и
// /aas/oauth2/te. Они оставлены для интеграций, где требуется или фактически
// работает подписание client_secret в формате PKCS#7, например через
// signature.LocalCryptoProPKCS7. PKCS#7-методы отличаются от основных методов не
// только адресом эндпоинта, но и составом строки, которая передается провайдеру
// подписи.
//
// # Проверка state
//
// Методы [Client.AuthURI] и [Client.AuthURIPKCS7] возвращают ссылку и state:
//
//	uri, state, err := client.AuthURI("openid", redirectURI, permissions)
//
// Приложение должно сохранить этот state на время пользовательского сценария
// и передать его в [Client.ParseCallback] при обработке callback-запроса:
//
//	code, _, err := client.ParseCallback(r.URL.Query(), state)
//
// Если state из callback-запроса не совпадает с ожидаемым значением,
// возвращается ErrStateMismatch. При обмене кода на маркер доступа и при
// обновлении маркера клиент также проверяет, что state в JSON-ответе ЕСИА
// совпадает со state отправленного token-запроса.
//
// # Примеры
//
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-request] — запрос согласия пользователя и получения маркера доступа
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-update] — обновление маркера доступа
//
// # Требования
//
// Системные требования:
//   - Go 1.21+
//   - Для подписания запросов к ЕСИА с помощью [signature.LocalCryptoPro] или
//     [signature.LocalCryptoProPKCS7] — КриптоПро CSP 5.0+ и сертификат для
//     подписания запросов
//
// Регламентные требования:
//  1. Информационная система (ИС) должна быть зарегистрирована на
//     Технологическом портале ЕСИА: продуктовом или тестовом (ТЕСИА)
//  2. Для ИС должен быть выпущен необходимый сертификат
//  3. Публичная часть сертификата должна быть загружена на Технологический портал ЕСИА
//  4. Выполнены все необходимые шаги регламента подключения ИС к тестовой
//     или продуктовой среде ЕСИА и согласована заявка на доступ ИС к необходимым скоупам
//
// # Особенности вычисления хэша сертификата
//
// Основные методы [Client.AuthURI], [Client.TokenExchange] и [Client.TokenUpdate]
// передают в ЕСИА параметр client_certificate_hash. По этому параметру ЕСИА
// выбирает сертификат ИС, которым должна проверяться подпись client_secret.
//
// Сертификат может храниться в PEM- или DER-представлении. В карточку ИС на
// Технологическом портале ЕСИА можно загрузить сертификат в любом из этих
// форматов.
//
// Для основных методов хэш client_certificate_hash должен быть вычислен от
// DER-представления сертификата. Если вычислить хэш от PEM-файла целиком, ЕСИА
// вернет ошибку ESIA-007053: OAuthErrorEnum.clientSecretWrong.
//
// PKCS#7-методы с PKCS#7-подписью могут не проявлять эту проблему на этапе
// формирования ссылки, поскольку сертификат передается внутри подписи
// client_secret, а параметр client_certificate_hash в PKCS#7-ссылке не
// используется.
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
// # Руководящие документы
//
//  1. Методические рекомендации по использованию ЕСИА: https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-ispolzovaniyu-esia
//  2. Методические рекомендации по интеграции с REST API Цифрового профиля: https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-integraczii-s-rest-api-czifrovogo-profilya
//  3. Сценарии использования инфраструктуры цифрового профиля физического лица: https://digital.gov.ru/documents/sczenarii-ispolzovaniya-infrastruktury-czifrovogo-profilya-fizicheskogo-licza
//  4. Регламент информационного взаимодействия ЕСИА: https://digital.gov.ru/documents/reglament-informaczionnogo-vzaimodejstviya-esia
//  5. Руководство пользователя технологического портала ЕСИА: https://digital.gov.ru/documents/rukovodstvo-polzovatelya-tehnologicheskogo-portala-esia
//  6. Руководство пользователя ЕСИА: https://digital.gov.ru/documents/rukovodstvo-polzovatelya-esia
//
// Ссылки на документы Минцифры могут изменяться. Если прямая ссылка недоступна,
// используйте каталог документов Минцифры или разделы справки Технологического
// портала ЕСИА.
//
// # Адреса Технологического портала ЕСИА
//   - Тестовая среда (ТЕСИА): https://esia-portal1.test.gosuslugi.ru/console/tech
//   - Продуктовая среда: https://esia.gosuslugi.ru/console/tech/
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
//
// # Страница предоставленных согласий пользователя на Портале Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-betalk.test.gosuslugi.ru/settings/third-party/agreements/acting
//   - Продуктовая среда: https://lk.gosuslugi.ru/settings/third-party/agreements/acting
package aas
