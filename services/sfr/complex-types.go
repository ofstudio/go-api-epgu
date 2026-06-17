package sfr

// AddressRus - российский адрес
type AddressRus struct {
	ZipCode    *string `xml:"ns2:Индекс"`
	Region     *string `xml:"ns2:РоссийскийАдрес>ns2:Регион>ns2:Название"`
	District   *string `xml:"ns2:РоссийскийАдрес>ns2:Район>ns2:Название"`
	City       *string `xml:"ns2:РоссийскийАдрес>ns2:Город>ns2:Название"`
	Settlement *string `xml:"ns2:РоссийскийАдрес>ns2:НаселенныйПункт>ns2:Название"`
	Street     *string `xml:"ns2:РоссийскийАдрес>ns2:Улица>ns2:Название"`
	House      *string `xml:"ns2:РоссийскийАдрес>ns2:Дом>ns2:Номер"`
	Housing    *string `xml:"ns2:РоссийскийАдрес>ns2:Корпус>ns2:Номер"`
	Building   *string `xml:"ns2:РоссийскийАдрес>ns2:Строение>ns2:Номер"`
	Flat       *string `xml:"ns2:РоссийскийАдрес>ns2:Квартира>ns2:Номер"`
}

// NewAddressRus - конструктор [AddressRus]
func NewAddressRus() *AddressRus {
	return &AddressRus{}
}

// WithZipCode - ns2:Индекс
func (a *AddressRus) WithZipCode(zipCode string) *AddressRus {
	a.ZipCode = &zipCode
	return a
}

// WithRegion - УТ:Регион
func (a *AddressRus) WithRegion(region string) *AddressRus {
	a.Region = &region
	return a
}

// WithDistrict - УТ:Район
func (a *AddressRus) WithDistrict(district string) *AddressRus {
	a.District = &district
	return a
}

// WithCity - УТ:Город
func (a *AddressRus) WithCity(city string) *AddressRus {
	a.City = &city
	return a
}

// Settlement - УТ:НаселенныйПункт
func (a *AddressRus) WithSettlement(settlement string) *AddressRus {
	a.Settlement = &settlement
	return a
}

// WithStreet - УТ:Улица
func (a *AddressRus) WithStreet(street string) *AddressRus {
	a.Street = &street
	return a
}

// WithHouse - УТ:Дом
func (a *AddressRus) WithHouse(house string) *AddressRus {
	a.House = &house
	return a
}

// WithHousing - УТ:Корпус
func (a *AddressRus) WithHousing(housing string) *AddressRus {
	a.Housing = &housing
	return a
}

// WithBuilding - УТ:Строение
func (a *AddressRus) WithBuilding(building string) *AddressRus {
	a.Building = &building
	return a
}

// WithFlat - УТ:Квартира
func (a *AddressRus) WithFlat(flat string) *AddressRus {
	a.Flat = &flat
	return a
}

// BirthPlace - УТ:МестоРождения
type BirthPlace struct {
	Type    string `xml:"ns2:ТипМестаРождения"`         // Пример: ОСОБОЕ
	City    string `xml:"ns2:ГородРождения,omitempty"`  // Пример: рп Михайловка, Ардатовский р-он
	Country string `xml:"ns2:СтранаРождения,omitempty"` // Пример: Российская Федерация
}

// ns2:МестоРождения/ns2:ТипМестаРождения
const BirthPlaceSpecial = "ОСОБОЕ"

// Citizenship - ns2:Гражданство/ns2:Тип
type CitizenshipType string

// ns2:Гражданство/ns2:Тип
const (
	CitizenshipRF        CitizenshipType = "1" // Гражданин РФ
	CitizenshipForeign   CitizenshipType = "2" // Иностранный гражданин
	CitizenshipStateless CitizenshipType = "3" // Лицо без гражданства
)

// Citizenship - ns2:Гражданство
type Citizenship struct {
	Type CitizenshipType `xml:"ns2:Тип"` // Пример: 1
}

// FIO - ns2:ФИО
type FIO struct {
	LastName       string `xml:"ns2:Фамилия"`            // Пример: ИВАНОВ
	FirstName      string `xml:"ns2:Имя"`                // Пример: ИВАН
	PatronymicName string `xml:"ns2:Отчество,omitempty"` // Пример: ИВАНОВИЧ
}

// IdentityDoc - УТ:УдостоверяющийДокументОграниченногоСрока
type IdentityDoc struct {
	Type       string `xml:"ns2:ТипДокумента"`               // Пример: ПАСПОРТ РОССИИ
	Series     string `xml:"ns2:Серия"`                      // Пример: 1234
	Number     string `xml:"ns2:Номер"`                      // Пример: 123456
	IssuedAt   Date   `xml:"ns2:ДатаВыдачи"`                 // Пример: 2010-04-13
	IssuedBy   string `xml:"ns2:КемВыдан"`                   // Пример: ОВД ЛЕНИНСКОГО РАЙОНА Г. САМАРЫ
	IssuerCode string `xml:"ns2:КодПодразделения,omitempty"` // Пример: 123456
}

// УТ:УдостоверяющийДокументОграниченногоСрока/УТ:ТипДокумента
const IdentityDocPassportRF = "ПАСПОРТ РОССИИ"
