package zird

import (
	"encoding/xml"

	"github.com/ofstudio/go-api-epgu/services/sfr"
)

// EDSFR — электронный документ заявления.
type EDSFR struct {
	XMLName xml.Name `xml:"ЭДСФР"`
	sfr.Namespaces
	ZIRD ZIRD `xml:"ЗИРД"`
}

// ZIRD — заявление о доставке пенсии через банк при личной подаче.
type ZIRD struct {
	TOSFR       TerritorialOffice `xml:"ТОСФР"`
	Applicant   Applicant         `xml:"Заявитель"`
	Delivery    Delivery          `xml:"Доставка"`
	FillingDate sfr.Date          `xml:"ДатаЗаполнения"`
}

// TerritorialOffice — территориальный орган СФР, выбранный вызывающим кодом.
type TerritorialOffice struct {
	Code string `xml:"ns2:Код"`
	Name string `xml:"ns2:Наименование"`
}

// Applicant — сведения о заявителе, который является получателем пенсии.
type Applicant struct {
	FIO             sfr.FIO      `xml:"ns2:ФИО"`
	SNILS           sfr.SNILS    `xml:"ns2:СНИЛС"`
	BirthDate       sfr.Date     `xml:"ns2:ДатаРождения"`
	Addresses       []AddressRus `xml:"Адрес"`
	DeliveryAddress *Address     `xml:"АдресДоставки,omitempty"`
	IdentityDoc     IdentityDoc  `xml:"ns2:УдостоверяющийДокумент"`
	Phone           string       `xml:"ns2:Телефон,omitempty"`
}

// AddressType — тип российского адреса.
type AddressType int

const (
	AddressRegistration AddressType = 1
	AddressResidence    AddressType = 2
	AddressActual       AddressType = 3
)

// AddressPart — название и необязательное сокращение части адреса.
type AddressPart struct {
	Name      string `xml:"ns2:Название"`
	ShortName string `xml:"ns2:Сокращение,omitempty"`
}

// Address — российский адрес в формате УТ/2025-01-01.
type Address struct {
	GUID         string       `xml:"ns2:GUID,omitempty"`
	ZipCode      string       `xml:"ns2:Индекс,omitempty"`
	Region       *AddressPart `xml:"ns2:Регион,omitempty"`
	Area         *AddressPart `xml:"ns2:Округ,omitempty"`
	District     *AddressPart `xml:"ns2:Район,omitempty"`
	City         *AddressPart `xml:"ns2:Город,omitempty"`
	CityDistrict *AddressPart `xml:"ns2:ВнутригородскойРайон,omitempty"`
	Settlement   *AddressPart `xml:"ns2:НаселенныйПункт,omitempty"`
	Street       *AddressPart `xml:"ns2:Улица,omitempty"`
	House        string       `xml:"ns2:Дом,omitempty"`
	Building     string       `xml:"ns2:Строение,omitempty"`
	Estate       string       `xml:"ns2:Владение,omitempty"`
	Housing      string       `xml:"ns2:Корпус,omitempty"`
	Flat         string       `xml:"ns2:Квартира,omitempty"`
}

// AddressRus — адрес регистрации, пребывания или фактического проживания.
type AddressRus struct {
	Type AddressType `xml:"ns2:ТипАдреса"`
	Address
}

// IdentityDoc — документ, удостоверяющий личность.
type IdentityDoc struct {
	Code       string   `xml:"ns2:Код"`
	Name       string   `xml:"ns2:Наименование"`
	Series     string   `xml:"ns2:Серия,omitempty"`
	Number     string   `xml:"ns2:Номер"`
	IssuedAt   sfr.Date `xml:"ns2:ДатаВыдачи"`
	IssuedBy   string   `xml:"ns2:КемВыдан"`
	IssuerCode string   `xml:"ns2:КодПодразделения,omitempty"`
}

// Delivery — доставка пенсии самому заявителю.
type Delivery struct {
	Pension PensionDelivery `xml:"ЗДП"`
}

func (d Delivery) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		Recipient int             `xml:"Получатель"`
		Pension   PensionDelivery `xml:"ЗДП"`
	}{1, d.Pension}, start)
}

// PensionDelivery — банковская доставка и запрос выплатного дела при смене адреса.
type PensionDelivery struct {
	RequestPaymentFile bool `xml:"ЗапросВыплатногоДела"` // Запросить выплатное дело при смене адреса доставки.
	Bank               Bank `xml:"Банк"`
}

// Bank — реквизиты российского банка из реестра СФР.
type Bank struct {
	Name          string `xml:"ns3:НаименованиеОрганизации"`
	BIC           string `xml:"ns3:БИК"`
	AccountNumber string `xml:"ns3:НомерСчета"`
	KeepPrevious  int    `xml:"ns3:УчестьПрежнююДоставку"` // 0 — не учитывать прежнюю дату доставки, 1 — учитывать.
}

func (b Bank) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		Name          string `xml:"ns3:НаименованиеОрганизации"`
		BIC           string `xml:"ns3:БИК"`
		AccountNumber string `xml:"ns3:НомерСчета"`
		KeepPrevious  int    `xml:"ns3:УчестьПрежнююДоставку"`
	}{b.Name, b.BIC, b.AccountNumber, b.KeepPrevious}, start)
}

// Request — транспортный конверт. RegNum в спецификации является опечаткой.
type Request struct {
	XMLName xml.Name `xml:"ns2:Request"`
	sfr.Namespaces
	AF                         string   `xml:"AF,attr"`
	SNILS                      string   `xml:"SNILS"`
	ExternalRegistrationNumber string   `xml:"ExternalRegistrationData>ExternalRegistrationNumber"`
	ExternalRegistrationDate   sfr.Date `xml:"ExternalRegistrationData>ExternalRegistrationDate"`
	MFCCode                    string   `xml:"MFCCode"`
	FRGUTargetId               string   `xml:"FRGUTargetId"`
	ApplicationFileName        string   `xml:"ApplicationFileName"`
}

var edsfrNamespaces = sfr.Namespaces{NS: "http://пф.рф/ЗИРД/2025-05-01", NS2: "http://пф.рф/УТ/2025-01-01", NS3: "http://пф.рф/ВЗЛ/типы/2025-01-01"}
var requestNamespaces = sfr.Namespaces{NS: "urn://iis.ecp.ru/zird/types/1.0.2", NS2: "urn://iis.ecp.ru/zird/1.0.2"}
