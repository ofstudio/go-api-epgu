package apipgu

import (
	"fmt"
)

type OrderMeta struct {
	Region      string
	ServiceCode string
	TargetCode  string
}

func (m *OrderMeta) JSON() []byte {
	return []byte(fmt.Sprintf(
		`{"region":"%s", "serviceCode":"%s", "targetCode":"%s"}`,
		m.Region, m.ServiceCode, m.TargetCode,
	))
}
