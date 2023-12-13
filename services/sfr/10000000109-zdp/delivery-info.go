package zdp

import "github.com/ofstudio/go-api-epgu/services/sfr"

// DeliveryLocation - место доставки пенсии из структуры [DeliveryInfo]
type DeliveryLocation int

const (
	DeliveryBankOrHome DeliveryLocation = 1 // Банк, доставка домой (в спецификации опечатка? Указано 3)
	DeliveryCashDesk   DeliveryLocation = 2 // В кассе Почты России, в кассе (в спецификации опечатка? Указано 4)
)

// DeliveryMethod - способ доставки пенсии из структуры [DeliveryInfo]
type DeliveryMethod int

const (
	DeliveryPostOffice DeliveryMethod = 1 // Почта России
	DeliveryBank       DeliveryMethod = 2 // Банк
	DeliveryOther      DeliveryMethod = 3 // Другая организация
)

// DeliveryRecipient - получатель пенсии из структуры [DeliveryInfo]
type DeliveryRecipient int

const (
	DeliveryMyself     DeliveryRecipient = 1 // Себе
	DeliveryMinorChild DeliveryRecipient = 2 // Несовершеннолетнему ребенку
)

// DeliveryPickup - способ вручения пенсии из структуры [DeliveryInfo]
type DeliveryPickup int

const (
	DeliveryOrganisation DeliveryPickup = 1 // В кассе Почты России, в кассе
	DeliveryHome         DeliveryPickup = 2 // Доставка домой
)

// DeliveryInfo - сведения о доставке пенсии из структуры [ZDP]
type DeliveryInfo struct {
	Date          sfr.Date          `xml:"ДатаДоставки"`                      // Пример: 2023-04-13
	Location      DeliveryLocation  `xml:"МестоДоставки"`                     // Пример: 1
	Method        DeliveryMethod    `xml:"СпособДоставки"`                    // Пример: 2
	Recipient     DeliveryRecipient `xml:"Получатель"`                        // Пример: 1
	Pickup        DeliveryPickup    `xml:"СпособВручения,omitempty"`          // Пример: 1
	Organisation  string            `xml:"НаименованиеОрганизации,omitempty"` // Пример: АЛТАЙСКИЙ РФ АО "РОССЕЛЬХОЗБАНК" г Барнаул
	AccountNumber string            `xml:"НомерСчета,omitempty"`              // Пример: 40817810000000000001
	Address       sfr.AddressRus    `xml:"Адрес"`                             // Адрес доставки
}
