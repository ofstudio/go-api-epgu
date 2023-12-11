package signature

// Provider - интерфейс провайдера электронной подписи запросов к ЕСИА.
// Провайдер должен реализовать
//   - подписание данных
//   - возвращать хэш сертификата
type Provider interface {
	Sign(data []byte) ([]byte, error)
	CertHash() string
}
