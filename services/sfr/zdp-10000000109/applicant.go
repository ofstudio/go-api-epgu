package zdp

import "github.com/ofstudio/go-api-epgu/services/sfr"

// Applicant - анкета заявителя структуры [ZDP]
type Applicant struct {
	FIO              sfr.FIO         `xml:"УТ:ФИО"`
	Sex              string          `xml:"УТ:Пол"`            // Пример: М
	BirthDate        sfr.Date        `xml:"УТ:ДатаРождения"`   // Пример: 1960-04-13
	SNILS            sfr.SNILS       `xml:"УТ:СтраховойНомер"` // Пример: 000-666-666 99
	BirthPlace       sfr.BirthPlace  `xml:"УТ:МестоРождения"`  // Место рождения
	Citizenship      sfr.Citizenship `xml:"УТ:Гражданство"`    // Пример: 1
	AddressFact      *sfr.AddressRus `xml:"УТ:АдресФактический,omitempty"`
	AddressReg       *sfr.AddressRus `xml:"УТ:АдресРегистрации,omitempty"`
	AddressResidence *sfr.AddressRus `xml:"УТ:АдресПребывания,omitempty"`
	Phone            string          `xml:"УТ:Телефоны>УТ:Телефон"`    // Пример: 89123456789
	Email            string          `xml:"УТ:АдресЭлПочты,omitempty"` // Пример: ivanov@mail.ru
	IdentityDoc      sfr.IdentityDoc `xml:"УТ:УдостоверяющийДокументОграниченногоСрока"`
}
