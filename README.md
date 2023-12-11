# go-api-epgu
[![Go Reference](https://pkg.go.dev/badge/github.com/ofstudio/go-api-epgu.svg)](https://pkg.go.dev/github.com/ofstudio/go-api-epgu)

REST-клиент для работы с [API Госуслуг (ЕПГУ)](https://partners.gosuslugi.ru/catalog/api_for_gu) 
и OAuth2-клиент для запроса согласия и маркера доступа ЕСИА
для получателей услуг ЕПГУ - физических лиц.

Библиотека поддерживают полное логирование запросов и ответов, а также обработку ошибок API 
и может быть использована для отладки взаимодействия с ЕПГУ и ЕСИА.

## REST-клиент для API Госуслуг

 - [Client.OrderCreate](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderCreate) - \
   создание заявления
 - [Client.OrderPushChunked](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderPushChunked) -
   загрузка архива по частям
 - [Client.OrderInfo](https://pkg.go.dev/github.com/ofstudio/go-api-epgu#Client.OrderInfo) -
   запрос детальной информации по отправленному заявлению

## Запрос согласия пользователя и получение маркера доступа ЕСИА

- [esia/aas](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas) — OAuth2-клиент для получения маркера доступа ЕСИА
- [esia/signature](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/signature) — подпись запросов к ЕСИА

## Примеры
- [Запрос согласия пользователя и получения маркера доступа](/examples/esia-token-request/main.go)
- [Обновление маркера доступа](/examples/esia-token-update/main.go)

## Установка

```
go get -u github.com/ofstudio/go-api-epgu
```
## Системные требования

- Go 1.21+
- Для подписания запросов к ЕСИА с помощью
  [LocalCryptoPro](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/signature#LocalCryptoPro) — 
  КриптоПро CSP 5.0+ и сертификат для подписания запросов 
     

## Регламентные требования
1. Информационная система должна быть зарегистрирована на Технологическом портале ЕСИА:
   продуктовом или тестовом (SVCDEV)
2. Для ИС должен быть выпущен необходимый сертификат
3. Публичная часть сертификата должна быть загружена на Технологический портал ЕСИА
4. Выполнены все необходимые шаги регламента и согласованы заявки на подключения ИС к тестовым
   или продуктовым средам ЕСИА и ЕПГУ

## Руководящие документы
1. [Портал API Госуслуг](https://partners.gosuslugi.ru/catalog/api_for_gu): 
   регламенты подключения, руководства, спецификация API ЕПГУ и отдельных услуг
2. [Методические рекомендации по интеграции с REST API Цифрового профиля](https://digital.gov.ru/ru/documents/7166/)
3. [Методические рекомендации по использованию ЕСИА](https://digital.gov.ru/ru/documents/6186/)
4. [Руководство пользователя ЕСИА](https://digital.gov.ru/ru/documents/6182/)
5. [Руководство пользователя технологического портала ЕСИА](https://digital.gov.ru/ru/documents/6190/)

## Ссылки

### ЕСИА
- Тестовая среда (SVCDEV): https://esia-portal1.test.gosuslugi.ru
- Продуктовая среда: https://esia.gosuslugi.ru

### Технологический портал ЕСИА
- Тестовая среда (SVCDEV): https://esia-portal1.test.gosuslugi.ru/console/tech
- Продуктовая среда: https://esia.gosuslugi.ru/console/tech/

### Портал Госуслуг
- Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
- Продуктовая среда: https://lk.gosuslugi.ru

### Список согласий предоставленных пользователем
- Тестовая среда (SVCDEV): https://svcdev-betalk.test.gosuslugi.ru/settings/third-party/agreements/acting
- Продуктовая среда: https://lk.gosuslugi.ru/settings/third-party/agreements/acting

## Лицензия
Распространяется по лицензии MIT. Более подробная информация в файле LICENSE.
