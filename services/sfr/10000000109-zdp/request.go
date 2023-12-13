package zdp

import (
	"encoding/xml"

	"github.com/ofstudio/go-api-epgu/services/sfr"
)

// Request - корневой элемент транспортного конверта заявления
type Request struct {
	XMLName xml.Name `xml:"ns2:Request"`
	sfr.Namespaces

	SNILS                      string   `xml:"SNILS"`                                               // Пример: 11702657331
	ExternalRegistrationNumber string   `xml:"ExternalRegistrationData>ExternalRegistrationNumber"` // Пример: 3500274591
	ExternalRegistrationDate   sfr.Date `xml:"ExternalRegistrationData>ExternalRegistrationDate"`   // Пример: 2023-04-13
	OKATO                      string   `xml:"OKATO"`                                               // Пример: 92401379000
	OKTMO                      string   `xml:"OKTMO"`                                               // Пример: 92701000001
	MFCCode                    string   `xml:"MFCCode"`                                             // Значение не проверяется
	ApplicationFileName        string   `xml:"ApplicationFileName"`                                 // Пример: req_2f1ee59c-a531-42f6-690e-c780dd2e345e.xml
	FRGUTargetId               string   `xml:"ns2:FRGUTargetId"`                                    // 10002953957
}

var requestNamespaces = sfr.Namespaces{
	NS:  "urn://cmv.pfr.ru/types/1.0.1",
	NS2: "urn://cmv.pfr.ru/zdp/1.0.1",
}
