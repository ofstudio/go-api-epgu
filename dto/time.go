package dto

import (
	"fmt"
	"strings"
	"time"
)

// "date": "2023-11-02T07:27:22.586+0300"
const apipguLayout = "2006-01-02T15:04:05.000-0700"

// Time - дата и время в формате API ЕПГУ.
//
//	2006-01-02T15:04:05.000-0700
type Time struct {
	time.Time
}

func (ct *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(apipguLayout, s)
	return
}

func (ct *Time) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, ct.Time.Format(apipguLayout))), nil
}
