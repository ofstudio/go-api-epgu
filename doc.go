// REST-клиент для API Госуслуг (АПИ ЕПГУ).
//
// https://partners.gosuslugi.ru/catalog/api_for_gu
//
// # Методы
//
//   - [Client.OrderCreate] — создание заявления
//   - [Client.OrderPushChunked] — загрузка архива по частям
//   - [Client.OrderInfo] — запрос детальной информации по отправленному заявлению
//
// # Получение маркера доступа (токена) ЕСИА
//
//   - [github.com/ofstudio/go-api-epgu/esia/aas] — OAuth2-клиент для работы с согласиями ЕСИА
//   - [github.com/ofstudio/go-api-epgu/esia/signature] — Электронная подпись запросов к ЕСИА
//
// # Примеры
//
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-request] — запрос согласия пользователя и получения маркера доступа
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-update] — обновление маркера доступа
//
// # Руководящие документы
//
//  1. Спецификация API, основные руководящие документы и регламенты подключения опубликованы на Портале API Госуслуг: https://partners.gosuslugi.ru/catalog/api_for_gu
//  2. Методические рекомендации по интеграции с REST API Цифрового профиля: https://digital.gov.ru/ru/documents/7166/
//  3. Методические рекомендации по использованию ЕСИА: https://digital.gov.ru/ru/documents/6186/
//  4. Руководство пользователя ЕСИА: https://digital.gov.ru/ru/documents/6182/
//  5. Руководство пользователя технологического портала ЕСИА: https://digital.gov.ru/ru/documents/6190/
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
package apipgu
