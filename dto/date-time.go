package dto

import (
	"fmt"
	"strings"
	"time"
)

// "date": "2023-11-02T07:27:22.586+0300"
const apipguLayout = "2006-01-02T15:04:05.000-0700"

// DateTime - дата и время в формате API ЕПГУ.
//
//	2006-01-02T15:04:05.000-0700
type DateTime struct {
	time.Time
}

func (d *DateTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		d.Time = time.Time{}
		return
	}
	d.Time, err = time.Parse(apipguLayout, s)
	return
}

func (d *DateTime) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format(apipguLayout))), nil
}
