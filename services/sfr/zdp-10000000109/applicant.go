package zdp

import "github.com/ofstudio/go-api-epgu/services/sfr"

// Applicant - анкета заявителя структуры [ZDP]
type Applicant struct {
	FIO              sfr.FIO         `xml:"ns2:ФИО"`
	Sex              string          `xml:"ns2:Пол"`            // Пример: М
	BirthDate        sfr.Date        `xml:"ns2:ДатаРождения"`   // Пример: 1960-04-13
	SNILS            sfr.SNILS       `xml:"ns2:СтраховойНомер"` // Пример: 000-666-666 99
	BirthPlace       sfr.BirthPlace  `xml:"ns2:МестоРождения"`  // Место рождения
	Citizenship      sfr.Citizenship `xml:"ns2:Гражданство"`    // Пример: 1
	AddressFact      *sfr.AddressRus `xml:"ns2:АдресФактический,omitempty"`
	AddressReg       *sfr.AddressRus `xml:"ns2:АдресРегистрации,omitempty"`
	AddressResidence *sfr.AddressRus `xml:"ns2:АдресПребывания,omitempty"`
	Phone            string          `xml:"ns2:Телефоны>ns2:Телефон"`   // Пример: 89123456789
	Email            string          `xml:"ns2:АдресЭлПочты,omitempty"` // Пример: ivanov@mail.ru
	IdentityDoc      sfr.IdentityDoc `xml:"ns2:УдостоверяющийДокументОграниченногоСрока"`
}
