package sfr

import (
	"encoding/xml"
	"fmt"
)

func ExampleNewDate() {
	type Example struct {
		Date Date `xml:"Date"`
	}
	doc := Example{Date: NewDate(2019, 1, 12)}

	result, err := xml.Marshal(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))

	// Output: <Example><Date>2019-01-12</Date></Example>
}

func ExampleNewDateTime() {
	type Example struct {
		DateTime DateTime `xml:"DateTime"`
	}
	doc := Example{DateTime: NewDateTime(2019, 1, 12, 13, 14, 15)}

	result, err := xml.Marshal(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))

	// Output: <Example><DateTime>2019-01-12T13:14:15</DateTime></Example>
}
