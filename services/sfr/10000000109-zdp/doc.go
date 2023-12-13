// Услуга "Доставка пенсии и социальных выплат ПФР"
//
// # Параметры услуги
//
//   - eServiceCode (код услуги): 10000000109
//   - serviceTargetCode (идентификатор цели услуги): -10000000109
//   - Идентификатор цели оказания госуслуги по ФРГУ: 10002953957
//   - Категории получателей: ФЛ с подтвержденной УЗ
//   - Подписание: не требуется
//   - Возможность отмены: не предусмотрена
//
// # URL формы
//
//   - Тестовая среда SVCDEV: https://svcdev-beta.test.gosuslugi.ru/600109/1/form
//   - Продуктивная среда: https://www.gosuslugi.ru/600109/1/form
//
// Отправка заявлений происходит с использованием вида сведений «Приём заявления о доставке пенсии»:
// https://lkuv.gosuslugi.ru/paip-portal/#/inquiries/card/63730133-ff80-11eb-ba23-33408f10c8dc
//
// # Примеры
//
//   - [github.com/ofstudio/go-api-epgu/examples/order-push-chunked] — создание заявления и загрузка архива по частям
//
// # Примечание
//
// Предназначено для демонстрации. Реализованы не все возможности услуги,
// а также отсутствуют проверки на полноту и валидность данных.
package zdp
