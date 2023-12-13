// Пример создания заявления и загрузки архива по частям.
//
// # Основные шаги процесса
//
//  1. Заполнение данных заявления
//  2. Создание заявления: POST /api/gusmev/order
//  3. Формирование архива с заявлением
//  4. Загрузка архива с заявлением: POST /api/gusmev/push/chunked
//
// В качестве примера используется услуга "Доставка пенсии и социальных выплат ПФР" (10000000109).
//
// # Требования
//  1. Выполнены все необходимые шаги регламента подключения ИС к тестовой
//     или продуктовой среде АПИ ЕПГУ и согласована заявка на доступ ИС
//     к необходимой услуге: https://partners.gosuslugi.ru/catalog/api_for_gu
//  2. Получен маркер доступа (токен) ЕСИА: см. пример [github.com/ofstudio/go-api-epgu/examples/esia-token-request].
//  3. Доступ к учетной записи пользователя на тестовом (SVCDEV) или продуктовом портале Госуслуг
//
// # Адреса Портала Госуслуг
//   - Тестовая среда (SVCDEV): https://svcdev-beta.test.gosuslugi.ru
//   - Продуктовая среда: https://lk.gosuslugi.ru
package main

import (
	"log"

	"github.com/ofstudio/go-api-epgu"
	"github.com/ofstudio/go-api-epgu/services/sfr/10000000109-zdp"
)

const (
	// Маркер доступа к API ЕПГУ
	accessToken = "<< access_token >>"

	// URL для отправки запросов к API ЕПГУ
	baseURI = "https://svcdev-beta.test.gosuslugi.ru"
)

func main() {
	// Создаем клиент для работы с API ЕПГУ
	apiClient := apipgu.
		NewClient(baseURI).
		WithDebug(log.Default()) // Включаем отладку

	// === ШАГ 1 ===
	// Создаем услугу и заполняем данные заявления
	srv, err := zdp.NewService(okato, oktmo, zdpData)
	if err != nil {
		log.Fatal(err)
	}
	// Включаем отладку: выводим в консоль все создаваемые XML-документы и метаданные услуги
	//srv.WithDebug(log.Default())

	// === ШАГ 2 ===
	// Создаем заявление: POST /api/gusmev/order
	orderId, err := apiClient.OrderCreate(accessToken, srv.Meta())
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Успешно создано заявление на ЕПГУ с номером ", orderId)

	// === ШАГ 3 ===
	// Формируем архив с заявлением
	archive, err := srv.Archive(orderId)
	if err != nil {
		log.Fatal(err)
	}

	// === ШАГ 4 ===
	// Загружаем архив с заявлением: POST /api/gusmev/push/chunked
	if err = apiClient.OrderPushChunked(accessToken, orderId, srv.Meta(), archive); err != nil {
		log.Fatal(err)
	}
	log.Print("Архив с вложениями успешно загружен на ЕПГУ в заявление с номером ", orderId)

	// === ШАГ 5 === (опциональный)
	// Сохраняем копию архива локально (для проверки)
	//filePath := "_data/" + archive.Name + ".zip"
	//if err = os.WriteFile(filePath, archive.Data, 0644); err != nil {
	//	log.Fatal(err)
	//}
	//log.Print("Архив с вложениями сохранен в файл ", filePath)

}
