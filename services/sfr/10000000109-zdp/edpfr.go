package zdp

import (
	"encoding/xml"

	"github.com/ofstudio/go-api-epgu/services/sfr"
)

// EDPFR - корневой элемент документа заявления
type EDPFR struct {
	XMLName xml.Name `xml:"ЭДПФР"`
	sfr.Namespaces
	ZDP         ZDP         `xml:"ЗДП"`
	ServiceInfo ServiceInfo `xml:"СлужебнаяИнформация"`
}

// ServiceInfo - служебная информация структуры [EDPFR]
type ServiceInfo struct {
	GUID                       string       `xml:"АФ:GUID"`      // Пример: 8f8b7e4b-dec8-4dac-8a02-3dcde44d4fb2
	DateTime                   sfr.DateTime `xml:"АФ:ДатаВремя"` // Пример: 2023-04-13T14:48:03
	ExternalRegistrationNumber string       `xml:"НомерВнешний"` // Пример: 2662455582
	ApplicationDate            sfr.Date     `xml:"ДатаПодачи"`   // Пример: 2023-04-13
}

var edpfrNamespaces = sfr.Namespaces{
	NS:  "http://пф.рф/ВЗЛ/ЗДП/2016-04-15",
	AF:  "http://пф.рф/АФ",
	UT:  "http://пф.рф/унифицированныеТипы/2014-01-01",
	VZL: "http://пф.рф/ВЗЛ/типы/2014-01-01",
}
