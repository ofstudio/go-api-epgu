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
	"github.com/ofstudio/go-api-epgu/services/sfr"
	"github.com/ofstudio/go-api-epgu/services/sfr/zdp-10000000109"
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
	if err = apiClient.OrderPushChunked(accessToken, orderId, archive); err != nil {
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

//
// Данные заявления
//

// ОКАТО, ОКТМО и адрес заявителя
var (
	okato       = "92000000000"
	oktmo       = "92000000000"
	addressData = sfr.NewAddressRus().
			WithZipCode("421001").
			WithRegion("Респ. Татарстан").
			WithCity("г. Казань").
			WithStreet("ул. Адоратского").
			WithHouse("д. 2А").
			WithHousing("корп. 1").
			WithFlat("кв. 1")
)

// Заявление на доставку пенсии
var zdpData = zdp.ZDP{
	TOSFR: "Клиентская служба (на правах отдела) в Ново-Савиновском районе г.Казани",
	Applicant: zdp.Applicant{
		FIO: sfr.FIO{
			LastName:       "ИВАНОВ",
			FirstName:      "ИВАН",
			PatronymicName: "ИВАНОВИЧ",
		},
		Sex:       "М",
		BirthDate: sfr.NewDate(1952, 10, 18),
		SNILS:     sfr.MustParseSNILS("787-900-175 50"),
		BirthPlace: sfr.BirthPlace{
			Type: sfr.BirthPlaceSpecial,
			City: "Г. ОРЕЛ",
		},
		Citizenship: sfr.Citizenship{Type: sfr.CitizenshipRF},
		AddressFact: addressData,
		Phone:       "89123456789",
		IdentityDoc: sfr.IdentityDoc{
			Type:     sfr.IdentityDocPassportRF,
			Series:   "1234",
			Number:   "567890",
			IssuedAt: sfr.NewDate(2000, 10, 20),
			IssuedBy: "ФМС России",
		},
	},
	DeliveryInfo: zdp.DeliveryInfo{
		Location:      zdp.DeliveryBankOrHome,
		Method:        zdp.DeliveryBank,
		Recipient:     zdp.DeliveryMyself,
		Organisation:  "Филиал Банка «Южный» в г. Казани",
		AccountNumber: "40817000000000000001",
		Address:       *addressData,
	},
	Confirmation: 1,
}
