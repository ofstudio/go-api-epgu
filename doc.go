// REST-клиент для API Госуслуг (АПИ ЕПГУ).
// Разработан в соответствии с документом "Спецификация API ЕПГУ, версия 1.12"
//
// https://partners.gosuslugi.ru/catalog/api_for_gu
//
// # Методы
//
//   - [Client.OrderCreate] — создание заявления
//   - [Client.OrderPushChunked] — загрузка архива по частям
//   - [Client.OrderPush] — формирование заявления единым методом
//   - [Client.OrderInfo] — запрос детальной информации по отправленному заявлению
//   - [Client.OrderCancel] — отмена заявления
//   - [Client.AttachmentDownload] — скачивание файла вложения созданного заявления
//   - [Client.Dict] — получение справочных данных
//
// # Получение маркера доступа (токена) ЕСИА
//
//   - [github.com/ofstudio/go-api-epgu/esia/aas] — OAuth2-клиент для работы с согласиями ЕСИА
//   - [github.com/ofstudio/go-api-epgu/esia/signature] — Электронная подпись запросов к ЕСИА
//
// # Услуги API ЕПГУ
//
//   - [github.com/ofstudio/go-api-epgu/services/sfr/10000000109-zdp] — Доставка пенсии и социальных выплат ПФР
//
// # Примеры
//
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-request] — запрос согласия пользователя и получения маркера доступа
//   - [github.com/ofstudio/go-api-epgu/examples/esia-token-update] — обновление маркера доступа
//   - [github.com/ofstudio/go-api-epgu/examples/order-push-chunked] — создание заявления и загрузка архива по частям
//   - [github.com/ofstudio/go-api-epgu/examples/order-info] — получение детальной информации по отправленному заявлению
//
// # Руководящие документы
//
//  1. Портал API Госуслуг — регламенты подключения, руководства, спецификация API ЕПГУ и отдельных услуг: https://partners.gosuslugi.ru/catalog/api_for_gu
//  2. Методические рекомендации по использованию ЕСИА: https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-ispolzovaniyu-esia
//  3. Методические рекомендации по интеграции с REST API Цифрового профиля: https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-integraczii-s-rest-api-czifrovogo-profilya
//  4. Сценарии использования инфраструктуры цифрового профиля физического лица: https://digital.gov.ru/documents/sczenarii-ispolzovaniya-infrastruktury-czifrovogo-profilya-fizicheskogo-licza
//  5. Регламент информационного взаимодействия ЕСИА: https://digital.gov.ru/documents/reglament-informaczionnogo-vzaimodejstviya-esia
//  6. Руководство пользователя технологического портала ЕСИА: https://digital.gov.ru/documents/rukovodstvo-polzovatelya-tehnologicheskogo-portala-esia
//  7. Руководство пользователя ЕСИА: https://digital.gov.ru/documents/rukovodstvo-polzovatelya-esia
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
package apipgu
