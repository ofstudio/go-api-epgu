package zdp

import "github.com/ofstudio/go-api-epgu/services/sfr"

// ZDP - данные заявления о доставке пенсии
type ZDP struct {
	TOSFR        string       `xml:"ВЗЛ:ТерОрган"`        // Пример: Клиентская служба в Ново-Савиновском районе Казани
	FillingDate  sfr.Date     `xml:"ВЗЛ:ДатаЗаполнения"`  // Пример: 2023-04-13
	Applicant    Applicant    `xml:"Анкета"`              // Анкета заявителя
	DeliveryInfo DeliveryInfo `xml:"СведенияОДоставке"`   // Сведения о доставке пенсии
	Confirmation int          `xml:"ПризнакОзнакомления"` // Пример: 1
}
