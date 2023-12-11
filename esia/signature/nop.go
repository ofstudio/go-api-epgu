package signature

import "errors"

// Nop - тестовый провайдер подписи запросов.
// Возвращает фиксированные значения подписи и хэша сертификата.
type Nop struct {
	signature string
	certHash  string
}

// NewNop - конструктор [Nop].
// Если в signature передана пустая строка, то при вызове [Nop.Sign] будет возвращена ошибка.
func NewNop(signature, certHash string) *Nop {
	return &Nop{signature: signature, certHash: certHash}
}

// Sign - возвращает фиксированную подпись.
// Если в конструкторе [NewNop] в качестве signature была передана пустая строка,
// то возвращается ошибка.
func (p *Nop) Sign(_ []byte) ([]byte, error) {
	if p.signature == "" {
		return nil, errors.New("test")
	}
	return []byte(p.signature), nil
}

// CertHash - возвращает фиксированный хэш сертификата.
func (p *Nop) CertHash() string {
	return p.certHash
}
