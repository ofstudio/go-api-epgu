package sfr

// AddressRus - российский адрес
type AddressRus struct {
	ZipCode  *string `xml:"УТ:Индекс"`
	Region   *string `xml:"УТ:РоссийскийАдрес>УТ:Регион>УТ:Название"`
	District *string `xml:"УТ:РоссийскийАдрес>УТ:Район>УТ:Название"`
	City     *string `xml:"УТ:РоссийскийАдрес>УТ:Город>УТ:Название"`
	Street   *string `xml:"УТ:РоссийскийАдрес>УТ:Улица>УТ:Название"`
	House    *string `xml:"УТ:РоссийскийАдрес>УТ:Дом>УТ:Номер"`
	Housing  *string `xml:"УТ:РоссийскийАдрес>УТ:Корпус>УТ:Номер"`
	Building *string `xml:"УТ:РоссийскийАдрес>УТ:Строение>УТ:Номер"`
	Flat     *string `xml:"УТ:РоссийскийАдрес>УТ:Квартира>УТ:Номер"`
}

// NewAddressRus - конструктор [AddressRus]
func NewAddressRus() *AddressRus {
	return &AddressRus{}
}

// WithZipCode - УТ:Индекс
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
	Type    string `xml:"УТ:ТипМестаРождения"`         // Пример: ОСОБОЕ
	City    string `xml:"УТ:ГородРождения,omitempty"`  // Пример: рп Михайловка, Ардатовский р-он
	Country string `xml:"УТ:СтранаРождения,omitempty"` // Пример: Российская Федерация
}

// УТ:МестоРождения/УТ:ТипМестаРождения
const BirthPlaceSpecial = "ОСОБОЕ"

// Citizenship - УТ:Гражданство/УТ:Тип
type CitizenshipType string

// УТ:Гражданство/УТ:Тип
const (
	CitizenshipRF        CitizenshipType = "1" // Гражданин РФ
	CitizenshipForeign   CitizenshipType = "2" // Иностранный гражданин
	CitizenshipStateless CitizenshipType = "3" // Лицо без гражданства
)

// Citizenship - УТ:Гражданство
type Citizenship struct {
	Type CitizenshipType `xml:"УТ:Тип"` // Пример: 1
}

// FIO - УТ:ФИО
type FIO struct {
	LastName       string `xml:"УТ:Фамилия"`            // Пример: ИВАНОВ
	FirstName      string `xml:"УТ:Имя"`                // Пример: ИВАН
	PatronymicName string `xml:"УТ:Отчество,omitempty"` // Пример: ИВАНОВИЧ
}

// IdentityDoc - УТ:УдостоверяющийДокументОграниченногоСрока
type IdentityDoc struct {
	Type       string `xml:"УТ:ТипДокумента"`               // Пример: ПАСПОРТ РОССИИ
	Series     string `xml:"УТ:Серия"`                      // Пример: 1234
	Number     string `xml:"УТ:Номер"`                      // Пример: 123456
	IssuedAt   Date   `xml:"УТ:ДатаВыдачи"`                 // Пример: 2010-04-13
	IssuedBy   string `xml:"УТ:КемВыдан"`                   // Пример: ОВД ЛЕНИНСКОГО РАЙОНА Г. САМАРЫ
	IssuerCode string `xml:"УТ:КодПодразделения,omitempty"` // Пример: 123456
}

// УТ:УдостоверяющийДокументОграниченногоСрока/УТ:ТипДокумента
const IdentityDocPassportRF = "ПАСПОРТ РОССИИ"
