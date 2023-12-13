package sfr

import (
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrSNILSFormat = errors.New("некорректный формат СНИЛС")
	ErrSNILSCheck  = errors.New("некорректная контрольная сумма СНИЛС")
)

// SNILS - УТ:СтраховойНомер
type SNILS struct {
	number string
}

var reSNILS = regexp.MustCompile(`^(\d{3})[\s-]?(\d{3})[\s-]?(\d{3})[\s-]?(\d{2})$`)

// ParseSNILS - анализирует строку с номером СНИЛС и возвращает [SNILS].
// Входной формат: 11 цифр, допускаются разделители пробелы и дефисы: "000-000-000 00".
// Возвращает ошибки:
//   - [ErrSNILSFormat] если строка не соответствует формату
//   - [ErrSNILSCheck] если контрольная сумма не совпадает
func ParseSNILS(number string) (SNILS, error) {
	matches := reSNILS.FindStringSubmatch(number)
	if len(matches) != 5 {
		return SNILS{}, ErrSNILSFormat
	}

	number = matches[1] + matches[2] + matches[3] + matches[4]
	if number[9:11] != snilsCheckSum(number[:9]) {
		return SNILS{}, ErrSNILSCheck
	}

	return SNILS{number: number}, nil
}

// MustParseSNILS вызывает [ParseSNILS]. В случае ошибки, завершается паникой.
func MustParseSNILS(number string) SNILS {
	snils, err := ParseSNILS(number)
	if err != nil {
		panic(err)
	}
	return snils
}

func snilsCheckSum(num string) string {
	sum := 0
	for i, d := range num {
		sum += int(d-'0') * (9 - i)

	}
	sum = sum % 101
	if sum == 100 {
		sum = 0
	}
	return fmt.Sprintf("%02d", sum)
}

// Number возвращает номер СНИЛС без разделителей.
func (s SNILS) Number() string {
	return s.number
}

// String возвращает СНИЛС в формате "000-000-000 00".
func (s SNILS) String() string {
	return fmt.Sprintf("%s-%s-%s %s", s.number[:3], s.number[3:6], s.number[6:9], s.number[9:])
}

// MarshalXML реализует интерфейс [xml.Marshaler] для типа [SNILS].
// Формат СНИЛС: "000-000-000 00".
func (s SNILS) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(s.String(), start)
}
