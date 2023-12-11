package apipgu

import (
	"fmt"
)

// OrderMeta - метаданные создаваемого заявления.
type OrderMeta struct {
	Region      string // Код интерактивной формы на ЕПГУ
	ServiceCode string // Код цели обращения услуги в ФРГУ
	TargetCode  string // Код ОКАТО местоположения пользователя (можно передавать код ОКАТО региона, если невозможно определить точнее)
}

// JSON - возвращает метаданные в формате JSON.
func (m *OrderMeta) JSON() []byte {
	return []byte(fmt.Sprintf(
		`{"region":"%s", "serviceCode":"%s", "targetCode":"%s"}`,
		m.Region, m.ServiceCode, m.TargetCode,
	))
}
