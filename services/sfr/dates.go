package sfr

import (
	"encoding/xml"
	"time"
)

// Date - дата в формате YYYY-MM-DD
type Date struct {
	time.Time
}

// NewDate - конструктор [Date].
func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// MarshalXML - реализация интерфейса [xml.Marshaler].
// Формат даты: YYYY-MM-DD.
func (d Date) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.Format("2006-01-02"), start)
}

// DateTime - дата и время в формате YYYY-MM-DDThh:mm:ss
type DateTime struct {
	time.Time
}

// NewDateTime - конструктор [DateTime].
func NewDateTime(year int, month time.Month, day, hour, min, sec int) DateTime {
	return DateTime{time.Date(year, month, day, hour, min, sec, 0, time.UTC)}
}

// MarshalXML - реализация интерфейса [xml.Marshaler].
// Формат даты и времени: YYYY-MM-DDThh:mm:ss.
func (d DateTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.Format("2006-01-02T15:04:05"), start)
}
