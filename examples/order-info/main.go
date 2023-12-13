// Пример получения детальной информации по отправленному заявлению.
//
// # Требования
//  1. Выполнены все необходимые шаги регламента подключения ИС к тестовой
//     или продуктовой среде АПИ ЕПГУ и согласована заявка на доступ ИС
//     к необходимой услуге: https://partners.gosuslugi.ru/catalog/api_for_gu
//  2. Получен маркер доступа (токен) ЕСИА: см. пример [github.com/ofstudio/go-api-epgu/examples/esia-token-request].
//  3. Доступ к учетной записи пользователя на тестовом (SVCDEV) или продуктовом портале Госуслуг
//  4. Номер отправленного заявления на ЕПГУ: см. пример [github.com/ofstudio/go-api-epgu/examples/order-push-chunked]
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
package main

import (
	"log"

	"github.com/ofstudio/go-api-epgu"
	"github.com/ofstudio/go-api-epgu/utils"
)

const (
	// Маркер доступа к API ЕПГУ
	accessToken = "<< access_token >>"

	// URL для отправки запросов к API ЕПГУ
	baseURI = "https://svcdev-beta.test.gosuslugi.ru"

	// Номер заявления на ЕПГУ
	orderId = 0000000000
)

func main() {
	// Создаем клиент для работы с API ЕПГУ
	apiClient := apipgu.
		NewClient(baseURI).
		WithDebug(log.Default()) // Включаем отладку

	// Запрашиваем детальную информацию по заявлению
	orderInfo, err := apiClient.OrderInfo(accessToken, orderId)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(utils.PrettyJSON(orderInfo))
}
