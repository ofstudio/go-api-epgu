# go-api-epgu/esia/aas

OAuth2-клиент для запроса согласия и маркера доступа ЕСИА
для получателей услуг ЕПГУ — физических лиц.

## Методы

- [Client.AuthURI](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas/#Client.AuthURI) — формирует ссылку на страницу ЕСИА для предоставления пользователем запрошенных прав
- [Client.ParseCallback](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas/#Client.ParseCallback) — возвращает код авторизации из callback-запроса к `redirect_uri`
- [Client.TokenExchange](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas/#Client.TokenExchange) — обменивает код авторизации на маркер доступа (токен)
- [Client.TokenUpdate](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/aas/#Client.TokenUpdate) — обновляет маркер доступа по идентификатору пользователя (OID)

## Примеры
- [Запрос согласия пользователя и получения маркера доступа](/examples/esia-token-request/main.go)
- [Обновление маркера доступа](/examples/esia-token-update/main.go)

## Системные требования

- Go 1.21+
- Для подписания запросов к ЕСИА с помощью
  [LocalCryptoPro](https://pkg.go.dev/github.com/ofstudio/go-api-epgu/esia/signature#LocalCryptoPro) —
  КриптоПро CSP 5.0+ и сертификат для подписания запросов

## Регламентные требования
1. Информационная система (ИС) должна быть зарегистрирована на
   Технологическом портале ЕСИА: продуктовом или тестовом (SVCDEV)
2. Для ИС должен быть выпущен необходимый сертификат
3. Публичная часть сертификата должна быть загружена на Технологический портал ЕСИА
4. Выполнены все необходимые шаги регламента подключения ИС к тестовой
   или продуктовой среде ЕСИА и согласована заявка на доступ ИС к необходимым скоупам

## Ссылки

### Руководящие документы
1. [Методические рекомендации по интеграции с REST API Цифрового профиля](https://digital.gov.ru/ru/documents/7166/)
2. [Методические рекомендации по использованию ЕСИА](https://digital.gov.ru/ru/documents/6186/)
3. [Руководство пользователя ЕСИА](https://digital.gov.ru/ru/documents/6182/)
4. [Руководство пользователя технологического портала ЕСИА](https://digital.gov.ru/ru/documents/6190/) 

### Адреса Технологического портала ЕСИА
- Тестовая среда (SVCDEV): https://esia-portal1.test.gosuslugi.ru/console/tech
- Продуктовая среда: https://esia.gosuslugi.ru/console/tech/

### Адреса Портала Госуслуг
- Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
- Продуктовая среда: https://lk.gosuslugi.ru

### Страница предоставленных согласий пользователя на Портале Госуслуг
- Тестовая среда (SVCDEV): https://svcdev-betalk.test.gosuslugi.ru/settings/third-party/agreements/acting
- Продуктовая среда: https://lk.gosuslugi.ru/settings/third-party/agreements/acting