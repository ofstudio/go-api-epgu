package zdp

import "github.com/ofstudio/go-api-epgu/services/sfr"

// ZDP - данные заявления о доставке пенсии
type ZDP struct {
	TOSFR        string       `xml:"ТерОрган"`                // Пример: Клиентская служба в Ново-Савиновском районе Казани
	FillingDate  sfr.Date     `xml:"ДатаЗаполнения"`          // Пример: 2023-04-13
	Applicant    Applicant    `xml:"ns3:Анкета"`              // Анкета заявителя
	DeliveryInfo DeliveryInfo `xml:"ns3:СведенияОДоставке"`   // Сведения о доставке пенсии
	Confirmation int          `xml:"ns3:ПризнакОзнакомления"` // Пример: 1
}
