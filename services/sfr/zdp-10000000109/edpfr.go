package zdp

import (
	"encoding/xml"

	"github.com/ofstudio/go-api-epgu/services/sfr"
)

// EDPFR - корневой элемент документа заявления
type EDPFR struct {
	XMLName xml.Name `xml:"ns3:ЭДПФР"`
	sfr.Namespaces
	ZDP         ZDP         `xml:"ns3:ЗДП"`
	ServiceInfo ServiceInfo `xml:"ns3:СлужебнаяИнформация"`
}

// ServiceInfo - служебная информация структуры [EDPFR]
type ServiceInfo struct {
	GUID                       string       `xml:"ns5:GUID"`         // Пример: 8f8b7e4b-dec8-4dac-8a02-3dcde44d4fb2
	DateTime                   sfr.DateTime `xml:"ns5:ДатаВремя"`    // Пример: 2023-04-13T14:48:03
	ExternalRegistrationNumber string       `xml:"ns3:НомерВнешний"` // Пример: 2662455582
	ApplicationDate            sfr.Date     `xml:"ns3:ДатаПодачи"`   // Пример: 2023-04-13
}

var edpfrNamespaces = sfr.Namespaces{
	NS:  "http://пф.рф/ВЗЛ/типы/2014-01-01",
	NS2: "http://пф.рф/унифицированныеТипы/2014-01-01",
	NS3: "http://пф.рф/ВЗЛ/ЗДП/2016-04-15",
	NS5: "http://пф.рф/АФ",
}
