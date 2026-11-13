# go-api-epgu

[![Go Reference](https://pkg.go.dev/badge/github.com/ofstudio/go-api-epgu.svg)](https://pkg.go.dev/github.com/ofstudio/go-api-epgu)
[![Go Report Card](https://goreportcard.com/badge/github.com/ofstudio/go-api-epgu)](https://goreportcard.com/report/github.com/ofstudio/go-api-epgu)

REST-клиент для работы с [API Госуслуг (ЕПГУ)](https://partners.gosuslugi.ru/catalog/api_for_gu).
Разработан в соответствии с документом [«Спецификация API ЕПГУ, версия 1.15»](/doc/Specifikaciya_API_EPGU_v1_15.docx)

## Методы

 - [Client.OrderCreate](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderCreate) — резервирование номера заявления
 - [Client.OrderPushChunked](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderPushChunked) — загрузка и отправка заявления после резервирования номера
 - [Client.OrderPushChunkedV2](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderPushChunkedV2) — загрузка и отправка заявления после резервирования номера с учетом дисклеймера
 - [Client.OrderPush](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderPush) — загрузка и отправка заявления одним запросом
 - [Client.OrderPushV2](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderPushV2) — загрузка и отправка заявления одним запросом с учетом дисклеймера
 - [Client.OrderInfo](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderInfo) — запрос детальной информации по отправленному заявлению
 - [Client.GetOrdersStatus](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.GetOrdersStatus) — получение статусов заявлений по списку заявлений
 - [Client.GetUpdatedAfter](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.GetUpdatedAfter) — получение статусов заявлений с даты обновления статуса
 - [Client.OrderCancel](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderCancel) — отмена заявления
 - [Client.AttachmentDownload](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.AttachmentDownload) — скачивание файла
 - [Client.Dict](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.Dict) — получение справочных данных

Методы V2 возвращают `*OrderResult`: номер заявления `OrderId` и сообщение о дисклеймере `Message`.
Предупреждение, например «Ведомственные работы. Сервис ведомства работает с перебоями»,
может сопровождать успешную отправку: в этом случае `err == nil`.
При отправке нескольких чанков в `Message` сохраняется последнее полученное непустое сообщение.
Отсутствующее или пустое сообщение следующего чанка не затирает предыдущее.

Если подача заявлений заблокирована, API возвращает код `service_unavailable`, например
с сообщением «Ведомственная авария. Попробуйте подать заявление позже».
Проверить эту ошибку можно через `errors.Is(err, apipgu.ErrCodeServiceUnavailable)`.

```go
result, err := client.OrderPushChunkedV2(ctx, accessToken, orderId, archive)
if err != nil {
    if errors.Is(err, apipgu.ErrCodeServiceUnavailable) {
        log.Printf("Подача заявлений временно недоступна: %v", err)
    } else {
        log.Printf("Ошибка отправки заявления: %v", err)
    }
    return
}
log.Printf("Заявление отправлено: %d", result.OrderId)
if result.Message != "" {
    log.Printf("Предупреждение: %s", result.Message)
}
```

## Запрос согласия и получение маркера доступа ЕСИА

- [esia/aas](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas) — OAuth2-клиент для получения маркера доступа ЕСИА
- [esia/signature](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/signature) — подпись запросов к ЕСИА

## Услуги API ЕПГУ

- [services/sfr/zdp-10000000109](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/services/sfr/zdp-10000000109) — Доставка пенсии и социальных выплат ПФР (10000000109, ЗДП)
- [services/sfr/zird-60013502](services/sfr/zird-60013502/README.md) — Прием заявлений о доставке пенсии и иных социальных выплат (60013502, ЗИРД)

## Примеры

- [Запрос согласия пользователя и получения маркера доступа](/examples/esia-token-request/main.go)
- [Обновление маркера доступа](/examples/esia-token-update/main.go)
- [Резервирование номера заявления и загрузка архива по частям](/examples/order-push-chunked/main.go)
- [Получение детальной информации по отправленному заявлению](/examples/order-info/main.go)

## Установка

```
go get -u github.com/ofstudio/go-api-epgu
```

## Системные требования

- Go 1.21+
- Для подписания запросов к ЕСИА с помощью [LocalCryptoPro](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/signature#LocalCryptoPro) — КриптоПро CSP 5.0+, ключ и сертификат для подписания запросов

## Регламентные требования

1. Информационная система должна быть зарегистрирована на Технологическом портале ЕСИА:
   продуктовом или тестовом (ТЕСИА)
2. Для ИС должен быть выпущен необходимый сертификат
3. Публичная часть сертификата должна быть загружена на Технологический портал ЕСИА
4. Выполнены все необходимые шаги регламента и согласованы заявки на подключения ИС к тестовым
   или продуктовым средам ЕСИА и ЕПГУ

## Руководящие документы

1. [Портал API Госуслуг](https://partners.gosuslugi.ru/catalog/api_for_gu):
   регламенты подключения, руководства, спецификация API ЕПГУ и отдельных услуг
2. [Методические рекомендации по использованию ЕСИА](https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-ispolzovaniyu-esia)
3. [Методические рекомендации по интеграции с REST API Цифрового профиля](https://digital.gov.ru/documents/metodicheskie-rekomendaczii-po-integraczii-s-rest-api-czifrovogo-profilya)
4. [Сценарии использования инфраструктуры цифрового профиля физического лица](https://digital.gov.ru/documents/sczenarii-ispolzovaniya-infrastruktury-czifrovogo-profilya-fizicheskogo-licza)
5. [Регламент информационного взаимодействия ЕСИА](https://digital.gov.ru/documents/reglament-informaczionnogo-vzaimodejstviya-esia)
6. [Руководство пользователя технологического портала ЕСИА](https://digital.gov.ru/documents/rukovodstvo-polzovatelya-tehnologicheskogo-portala-esia)
7. [Руководство пользователя ЕСИА](https://digital.gov.ru/documents/rukovodstvo-polzovatelya-esia)
8. [Документы: Межведомственное электронное взаимодействие](https://info.gosuslugi.ru/docs/section/СМЭВ/)
9. [Документы: СМЭВ4](https://info.gosuslugi.ru/docs/section/СМЭВ/Методические_документы/СМЭВ4/):
   регламенты, руководство администратора, дистрибутив Агента ПОДД
10. [Документы: ПО для Участников СМЭВ](https://info.gosuslugi.ru/docs/section/ПО_для_СМЭВ/)
11. [Обмен в СМЭВ4 c использованием REST-сервиса](https://info.gosuslugi.ru/articles/Обмен_в_СМЭВ4_c_использованием_REST-сервиса/)

## Ссылки

### ЕСИА

- Тестовая среда (ТЕСИА): https://esia-portal1.test.gosuslugi.ru
- Продуктовая среда: https://esia.gosuslugi.ru

### Технологический портал ЕСИА

- Тестовая среда (ТЕСИА): https://esia-portal1.test.gosuslugi.ru/console/tech
- Продуктовая среда: https://esia.gosuslugi.ru/console/tech/

### Список согласий предоставленных пользователем

- Тестовая среда (SVCDEV): https://svcdev-betalk.test.gosuslugi.ru/settings/third-party/agreements/acting
- Продуктовая среда: https://lk.gosuslugi.ru/settings/third-party/agreements/acting

### Подключение

#### Подключение через TLS

Прямое подключение к API ЕПГУ через TLS-соединение.
Подробнее см "Спецификация API ЕПГУ версия 1.15", раздел "1.2. Реализация подключения по ГОСТ TLS"

- Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
- Продуктовая среда: https://lk.gosuslugi.ru

#### Подключение через СМЭВ4 (Агент ПОДД)

Подключение через регламентированный запрос типа REST-сервис в среде СМЭВ4.
Подробнее см "Спецификация API ЕПГУ версия 1.15", раздел "1.3. Подключение через СМЭВ 4".

- Тестовая среда (SVCDEV): https://nsud.gosuslugi.ru/pud/paip-portal/#/podd/open-api/specifications/card/e28f1ae0-0fdc-431a-9adb-17173564d1db
- Продуктовая среда: https://nsud.gosuslugi.ru/pud/paip-portal/#/podd/open-api/specifications/card/936fbb29-07bb-4206-8374-395eb2bd36b0

## Лицензия

Распространяется по лицензии MIT. Более подробная информация в файле LICENSE.
