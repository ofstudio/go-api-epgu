package apipgu

import (
	"encoding/json"
)

// OrderMeta - метаданные создаваемого заявления.
type OrderMeta struct {
	Region      string // Код ОКАТО местоположения пользователя (можно передавать код ОКАТО региона, если невозможно определить точнее)
	ServiceCode string // Код интерактивной формы на ЕПГУ
	TargetCode  string // Код цели обращения услуги в ФРГУ
}

// JSON - возвращает метаданные в формате JSON.
func (m *OrderMeta) JSON() []byte {
	body, _ := json.Marshal(struct {
		Region      string `json:"region"`
		ServiceCode string `json:"serviceCode"`
		TargetCode  string `json:"targetCode"`
	}{
		Region:      m.Region,
		ServiceCode: m.ServiceCode,
		TargetCode:  m.TargetCode,
	})
	return body
}
