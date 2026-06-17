package sfr

import (
	"encoding/xml"
	"fmt"
)

func ExampleNewAddressRus() {
	addr := NewAddressRus().
		WithZipCode("397140").
		WithRegion("обл Тамбовская").
		WithCity("г Тамбов").
		WithStreet("ул Советская").
		WithHouse("д 10").
		WithBuilding("стр 1").
		WithFlat("кв 1")

	xmlBytes, err := xml.MarshalIndent(addr, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(xmlBytes))

	// Output: <AddressRus>
	//   <ns2:Индекс>397140</ns2:Индекс>
	//   <ns2:РоссийскийАдрес>
	//     <ns2:Регион>
	//       <ns2:Название>обл Тамбовская</ns2:Название>
	//     </ns2:Регион>
	//     <ns2:Город>
	//       <ns2:Название>г Тамбов</ns2:Название>
	//     </ns2:Город>
	//     <ns2:Улица>
	//       <ns2:Название>ул Советская</ns2:Название>
	//     </ns2:Улица>
	//     <ns2:Дом>
	//       <ns2:Номер>д 10</ns2:Номер>
	//     </ns2:Дом>
	//     <ns2:Строение>
	//       <ns2:Номер>стр 1</ns2:Номер>
	//     </ns2:Строение>
	//     <ns2:Квартира>
	//       <ns2:Номер>кв 1</ns2:Номер>
	//     </ns2:Квартира>
	//   </ns2:РоссийскийАдрес>
	// </AddressRus>
}
