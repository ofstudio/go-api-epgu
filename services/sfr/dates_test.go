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
