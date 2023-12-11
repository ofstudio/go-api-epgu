package apipgu

import "time"

// OrderInfo - детальная информация по отправленному заявлению.
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.4. Получение деталей по заявлению".
//
// # Пример
//
// Пример для заявления "Доставка пенсии и социальных выплат СФР" (10000000109):
//
//	{
//	  "code": "OK",
//	  "message": null,
//	  "messageId": "2252fb21-92f8-61ee-a6f0-7ed53c117861",
//	  "order": {...}
//	}
type OrderInfo struct {
	Code      string // Код состояния заявления в соответствии с Приложением 1 Спецификации
	Message   string // Текстовое сообщение, описывающее текущее состояние запроса на создание заявления
	MessageId string // [Не документировано, GUID]
	Order     *Order // Детали заявления, если оно уже создано на портале и отправлено в ведомство
}

// Order - детальная информация по заявлению (см  [OrderInfo]).
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
//
// Примечание: поля, отмеченные как не документированные, не описаны
// в спецификации, однако могут приходить в ответе.
//
// # Пример
//
// Пример для заявления "Доставка пенсии и социальных выплат СФР" (10000000109).
// Обратите внимание, что структура в примере содержит не все все поля, упомянутые в спецификации.
//
//	{
//	 ////
//	 // Основные аттрибуты
//	 ////
//
//	 "id": 1230254874,         // Номер заявления
//	 "orderStatusId": 2,       // Код статуса заявления
//	 "statuses": [             // Статусы заявления
//	   {
//	     "id": 12300714241,                        // Идентификатор статуса
//	     "statusId": 0,                            // Код статуса
//	     "title": "Черновик заявления",            // Наименование статуса
//	     "date": "2023-11-02T07:27:22.586+0300",   // Дата и время смены статуса
//	     "orderId": 1230254874,                    // Номер заявления
//	     "finalStatus": false,                     // Флаг финального статуса
//	     "cancelAllowed": false,                   // Флаг возможности отменить заявление
//	     "hasResult": "N",                         // Флаг передачи файла в ответ на заявление
//	     "unreadEvent": true,                      // Признак прочтения
//	     "deliveryCancelAllowed": false,           // Флаг наличия отмены доставки
//	     "sendMessageAllowed": false,              // Признак разрешения отправки сообщения
//	     "editAllowed": false,                     // Признак редактирования
//	     "statusColorCode": "edit"                 // [Не документировано]
//	   },
//	   {
//	     "id": 12300712489,
//	     "statusId": 17,
//	     "title": "Зарегистрировано на портале",
//	     "date": "2023-11-02T07:27:22.936+0300",
//	     "orderId": 1230254874,
//	     "finalStatus": false,
//	     "hasResult": "N",
//	     "cancelAllowed": false,
//	     "sender": "Фонд пенсионного и социального страхования Российской Федерации", // Отправитель СМЭВ-сообщения о смене статуса
//	     "unreadEvent": true,
//	     "deliveryCancelAllowed": false,
//	     "sendMessageAllowed": false,
//	     "editAllowed": false,
//	     "statusColorCode": "in_progress"
//	   },
//	   {
//	     "id": 12300710521,
//	     "statusId": 21,
//	     "title": "Заявление отправлено в ведомство",
//	     "date": "2023-11-02T07:27:23.527+0300",
//	     "orderId": 1230254874,
//	     "finalStatus": false,
//	     "hasResult": "N",
//	     "cancelAllowed": false,
//	     "sender": "Фонд пенсионного и социального страхования Российской Федерации",
//	     "unreadEvent": true,
//	     "deliveryCancelAllowed": false,
//	     "sendMessageAllowed": false,
//	     "editAllowed": false,
//	     "statusColorCode": "in_progress",
//	   },
//	   {
//	     "id": 12300710522,
//	     "statusId": 2,
//	     "title": "Заявление получено ведомством",
//	     "date": "2023-11-02T07:27:44.134+0300",
//	     "orderId": 1230254874,
//	     "finalStatus": false,
//	     "hasResult": "N",
//	     "cancelAllowed": false,
//	     "sender": "Фонд пенсионного и социального страхования Российской Федерации",
//	     "comment": "Сообщение доставлено", // Комментарий к статусу
//	     "unreadEvent": true,
//	     "deliveryCancelAllowed": false,
//	     "sendMessageAllowed": false,
//	     "editAllowed": false,
//	     "statusColorCode": "in_progress"
//	   }
//	 ],
//	 "currentStatusHistory": { // История статуса
//	   "id": 12300710522,                        // Идентификатор статуса
//	   "statusId": 2,                            // Код статуса
//	   "title": "Заявление получено ведомством", // Код статуса
//	   "date": "2023-11-02T07:27:44.134+0300",   // Дата и время смены статуса
//	   "orderId": 1230254874,                    // Номер заявления
//	   "finalStatus": false,                     // Флаг финального статуса
//	   "hasResult": "N",                         // Флаг передачи файла в ответ на заявление
//	   "cancelAllowed": false,                   // Флаг наличия отмены
//	   "sender": "Фонд пенсионного и социального страхования Российской Федерации", // Наименование ведомства
//	   "comment": "Сообщение доставлено",        // Комментарий
//	   "unreadEvent": true,                      // Признак прочтение события
//	   "deliveryCancelAllowed": false,           // Флаг наличия отмены доставки
//	   "sendMessageAllowed": false,              // Признак разрешения отправки сообщения
//	   "editAllowed": false,                     // Признак редактирования
//	   "statusColorCode": "in_progress",         // [Не документировано]
//	 },
//	 "updated": "2023-11-02T07:27:44.140+0300", // Дата и время обновления статуса заявления
//	 "closed": false,          // Флаг наличия финального статуса
//	 "hasResult": false,       // Флаг передачи файла в ответ на заявление
//	 "orderAttachmentFiles": [ // Файлы заявления, отправленные пользователем
//	   {
//	     "id": "1230254874/files/mzXxRzhkODcwOWRiLWRkNDUtNDEyOS1hZTMyLTZiNGNlZmVjYTkwYy54bWw", // Идентификатор файла
//	     "fileName": "req_8d8567db-d445-4759-a122-6b4cefeca22c.xml",                           // Название файла
//	     "mimeType": "application/xml",                                                        // MIME-тип
//	     "link": "terrabyte://00/1230254874/req_8d8567db-d445-4759-a122-6b4cefeca22c.xml/2",   // Ссылка на файл в хранилище
//	     "hasDigitalSignature": false,                                                         // Наличие подписи
//	     "fileSize": 5519,                                                                     // Наличие подписи
//	     "type": "REQUEST"                                                                     // Наличие подписи
//	   },
//	   {
//	     "id": "1230254874/files/dHJhbnNYTRQ4NzA5ZGItZGQ0NS95MTI5LWFlMzItNmI0Y2VmZWNhOTBjLnhtbA",
//	     "fileName": "trans_8d8567db-d445-4759-a122-6b4cefeca22c.xml",
//	     "mimeType": "application/xml",
//	     "link": "terrabyte://00/1230254874/trans_8d8567db-d445-4759-a122-6b4cefeca22c.xml/2",
//	     "hasDigitalSignature": false,
//	     "fileSize": 644,
//	     "type": "ATTACHMENT"
//	   }
//	 ],
//	 "orderResponseFiles": [], // Информация о файлах в ответе заявления
//
//	 ////
//	 // Дополнительные аттрибуты
//	 ////
//
//	 "hasNewStatus": true,                               // Флаг нового статуса для заявления
//	 "currentStatusHistoryId": 12300710522,              // Идентификатор статуса заявления
//	 "orderStatusName": "Заявление получено ведомством", // Наименование статуса заявления
//
//	 "stateOrgId": 266,                 // Код ведомства
//	 "stateStructureName": "СФР",       // Наименование ведомства
//	 "stateOrgCode": "pfr",             // Сокращенное наименование ведомства
//	 "stateStructureId": "10000002796", // Код ведомства [по ФРГУ]
//	 "gisdo": false,                    // Признак подключенности ведомства к ФГИС ДО
//
//	 "sourceSystem": "Банк ЮЖНЫЙ",        // Наименование системы откуда было подано заявление [мнемоника ИС-потребителя API ЕПГУ]
//	 "creationMode": "api",               // Режим создания
//	 "extSystem": false,                  // Признак, что создано внешней системой (через сервис ЕЛК)
//	 "ownerId": 1000572618,               // Идентификатор пользователя [OID на Госуслугах / ЕСИА]
//	 "userId": 1000572618,                // Идентификатор пользователя [OID на Госуслугах / ЕСИА]
//	 "personType": "PERSON",              // Тип пользователя
//	 "userSelectedRegion": "00000000000", // Код ОКАТО местоположения пользователя
//	 "testUser": false,                   // Флаг тестового пользователя
//	 "location": "92000000000",           // Код уровня услуги [ОКАТО пользователя?]
//
//	 "orderType": "ORDER",               // Тип заявления
//	 "eserviceId": "10000000109",        // Идентификатор формы заявления
//	 "serviceTargetId": "-10000000109",  // Идентификатор цели
//	 "servicePassportId": "600109",      // Идентификатор паспорта услуги
//	 "serviceName": "Доставка пенсии и социальных выплат СФР", // Наименование цели
//	 "deprecatedService": false,        // Признак, что услуга больше не заказывается
//	 "hubForm": false,                  // Признак, что форма-концентратор
//	 "admLevelCode": "FEDERAL",         // Уровень услуги (региональный/федеральный)
//	 "multRegion": true,                // Признак регионозависимости
//	 "serviceEpguId": "1",              // Идентификатор цели услуги ЕПГУ
//	 "formVersion": "1",                // Версия
//	 "possibleServices": {},            // [Не документировано]
//
//	 "orderDate": "2023-11-02T07:27:22.000+0300",    // Дата и время создания заявления
//	 "requestDate": "2023-11-02T07:27:22.942+0300",  // Метка даты и времени запроса
//	 "orderAttributeEvents": [],       // Атрибуты событий для заявления
//	 "online": false,                  // Признак, онлайн услуга или нет
//	 "hasTimestamp": false,            // Флаг timestamp
//	 "hasActiveInviteToEqueue": false, // Флаг записи на прием
//	 "hasChildren": false,             // Флаг наличия дочерних заявлений
//	 "hasPreviewPdf": false,           // Флаг наличия пдф
//	 "hasEmpowerment2021": false,      // Флаг наличия делегирования
//	 "allowToEdit": false,             // Флаг редактирования заявления
//	 "allowToDelete": false,           // Флаг удаления заявки
//	 "draftHidden": false,             // Признак скрытия черновика
//	 "checkQueue": false,              // Флаг проверки очереди
//	 "eQueueEvents": [],               // Массив объектов eQueueEvent [структура элемента массива не документирована]
//	 "useAsTemplate": false,           // Флаг черновика заявления
//	 "withDelivery": false,            // Флаг доставки
//	 "withCustomResult": false,        // Признак необходимости отображения кнопки в Деталях заявления услуги
//	 "readyToPush": false,             // Служебный параметр
//	 "elk": false,                     // [Не документировано]
//
//	 "smevTx": "e74bc34c-c156-8523-1234-e6c549a28e23", // Код транзакции СМЭВ3
//	 "smevMessageId": "WAIT_RESPONSE",                 // Идентификатор СМЭВ-сообщения от ведомства, сменившего статус
//
//	 "paymentRequired": false,     // Флаг наличия оплаты
//	 "noPaidPaymentCount": -1,     // Количество неоплаченных платежей
//	 "paymentCount": 0,            // Количество платежей
//	 "hasNoPaidPayment": false,    // Флаг наличия оплаченного платежа
//	 "paymentStatusEvents": [],    // Статус событий при оплате [структура события оплаты не документирована]
//	 "orderPayments": [],          // Информация о платежах [структура объекта платежа не документирована]
//	 "payback": false,             // Служебный параметр
//
//	 "readyToSign": false,               // Для ЮЛ, для подписания заявки, Маркер ожидания УКЭП
//	 "signCnt": 0,                       // Кол-во подписей, для заявлений от нескольких заявителей
//	 "allFileSign": false,               // Флаг наличия ЭП для файлов
//	 "childrenSigned": false,            // Флаг подписи дочерних заявлений
//	 "edsStatus": "EDS_NOT_SUPPORTED",   // Идентификатор статуса проверки ЭП
//
//	 "infoMessages": [],       // Информация о сообщениях [структура объекта сообщения не документирована]
//	 "textMessages": [],       // Информация о сообщениях [структура объекта текстового сообщения не документирована]
//	 "unreadMessageCnt": 0,    // unreadMessageCnt
//
//	 "qrlink": { // [Не документировано]
//	   "hasAltMimeType": false,        // Связанно с alternativeMimeTypes из сервиса тербайта
//	   "fileSize": 0,                  // Размер файла
//	   "hasDigitalSignature": false,   // Флаг наличия ЭП
//	   "canSentToMFC": false,          // Флаг отправки в МФЦ
//	   "canPrintMFC": false            // [Не документировано]
//	 },
//	 "steps": [] // [Не документировано]
//	}
type Order struct {

	// Основные аттрибуты

	Id                   int                   `json:"id"`                   // Номер заявления
	OrderStatusId        int                   `json:"orderStatusId"`        // Код статуса заявления
	Statuses             []OrderStatus         `json:"statuses"`             // Статусы заявления
	CurrentStatusHistory []OrderStatusHistory  `json:"currentStatusHistory"` // История статуса
	Updated              time.Time             `json:"updated"`              // Дата и время обновления статуса заявления todo FORMAT!
	Closed               bool                  `json:"closed"`               // Флаг наличия финального статуса
	HasResult            bool                  `json:"hasResult"`            // Флаг передачи файла в ответ на заявление
	OrderAttachmentFiles []OrderAttachmentFile `json:"orderAttachmentFiles"` // Файлы заявления, отправленные пользователем
	OrderResponseFiles   []OrderResponseFile   `json:"orderResponseFiles"`   // Информация о файлах в ответе заявления

	// Дополнительные аттрибуты

	HasNewStatus           bool   `json:"hasNewStatus"`           // Флаг нового статуса для заявления
	CurrentStatusHistoryId int    `json:"currentStatusHistoryId"` // Идентификатор статуса заявления
	OrderStatusName        string `json:"orderStatusName"`        // Наименование статуса заявления
	StateOrgStatusCode     string `json:"stateOrgStatusCode"`     // Код ведомственного статуса
	StateOrgStatusName     string `json:"stateOrgStatusName"`     // Наименование ведомственного статуса

	StateOrgId         int    `json:"stateOrgId"`         // Код ведомства
	StateStructureName string `json:"stateStructureName"` // Наименование ведомства
	StateOrgCode       string `json:"stateOrgCode"`       // Сокращенное наименование ведомства
	StateStructureId   string `json:"stateStructureId"`   // Код ведомства [по ФРГУ]
	Gisdo              bool   `json:"gisdo"`              // Признак подключенности ведомства к ФГИС ДО

	SourceSystem       string `json:"sourceSystem"`       // Наименование системы откуда было подано заявление [мнемоника ИС-потребителя API ЕПГУ]
	CreationMode       string `json:"creationMode"`       // Режим создания
	ExtSystem          bool   `json:"extSystem"`          // Признак, что создано внешней системой (через сервис ЕЛК)
	OwnerId            int    `json:"ownerId"`            // Идентификатор пользователя [OID на Госуслугах / ЕСИА]
	UserId             int    `json:"userId"`             // Идентификатор пользователя [OID на Госуслугах / ЕСИА]
	PersonType         string `json:"personType"`         // Тип пользователя
	UserSelectedRegion string `json:"userSelectedRegion"` // Код ОКАТО местоположения пользователя
	TestUser           bool   `json:"testUser"`           // Флаг тестового пользователя
	Location           string `json:"location"`           // Код уровня услуги [ОКАТО пользователя?]
	OrgUserName        string `json:"orgUserName"`        // Наименование организации пользователя

	OrderType         string         `json:"orderType"`         // Тип заявления
	EserviceId        string         `json:"eserviceId"`        // Идентификатор формы заявления
	ServiceTargetId   string         `json:"serviceTargetId"`   // Идентификатор цели
	ServicePassportId string         `json:"servicePassportId"` // Идентификатор паспорта услуги
	ServiceName       string         `json:"serviceName"`       // Наименование цели
	DeprecatedService bool           `json:"deprecatedService"` // Признак, что услуга больше не заказывается
	HubForm           bool           `json:"hubForm"`           // Признак, что форма-концентратор
	HubFormVersion    int            `json:"hubFormVersion"`    // Идентификатор регионо-зависимой формы старого конструктора форм
	AdmLevelCode      string         `json:"admLevelCode"`      // Уровень услуги (региональный/федеральный)
	MultRegion        bool           `json:"multRegion"`        // Признак регионозависимости
	ServiceEpguId     string         `json:"serviceEpguId"`     // Идентификатор цели услуги ЕПГУ
	FormVersion       string         `json:"formVersion"`       // Версия
	ServiceUrl        string         `json:"serviceUrl"`        // Ссылка на заявление
	PortalCode        string         `json:"portalCode"`        // Код портала
	PortalName        string         `json:"portalName"`        // Наименование портала
	PossibleServices  map[string]any `json:"possibleServices"`  // [Не документировано]

	OrderDate               time.Time             `json:"orderDate"`               // Дата и время создания заявления TODO Format
	RequestDate             time.Time             `json:"requestDate"`             // Метка даты и времени запроса TODO Format
	OrderAttributeEvents    []OrderAttributeEvent `json:"orderAttributeEvents"`    // Атрибуты событий для заявления
	Online                  bool                  `json:"online"`                  // Признак, онлайн услуга или нет
	HasTimestamp            bool                  `json:"hasTimestamp"`            // Флаг timestamp
	HasActiveInviteToEqueue bool                  `json:"hasActiveInviteToEqueue"` // Флаг записи на прием
	HasChildren             bool                  `json:"hasChildren"`             // Флаг наличия дочерних заявлений
	HasPreviewPdf           bool                  `json:"hasPreviewPdf"`           // Флаг наличия пдф
	HasEmpowerment2021      bool                  `json:"hasEmpowerment2021"`      // Флаг наличия делегирования
	AllowToEdit             bool                  `json:"allowToEdit"`             // Флаг редактирования заявления
	AllowToDelete           bool                  `json:"allowToDelete"`           // Флаг удаления заявки
	DraftHidden             bool                  `json:"draftHidden"`             // Признак скрытия черновика
	CheckQueue              bool                  `json:"checkQueue"`              // Флаг проверки очереди
	EQueueEvents            []map[string]any      `json:"EQueueEvents"`            // Массив объектов eQueueEvent [структура элемента массива не документирована]
	UseAsTemplate           bool                  `json:"useAsTemplate"`           // Флаг черновика заявления
	WithDelivery            bool                  `json:"withDelivery"`            // Флаг доставки
	WithCustomResult        bool                  `json:"withCustomResult"`        // Признак необходимости отображения кнопки в Деталях заявления услуги
	PowerMnemonic           string                `json:"powerMnemonic"`           // Мнемоника полномочия, с которым подается заявление
	ReadyToPush             bool                  `json:"readyToPush"`             // Служебный параметр
	Elk                     bool                  `json:"elk"`                     // [Не документировано]

	SmevTx        string `json:"smevTx"`        // Код транзакции СМЭВ3
	SmevMessageId string `json:"smevMessageId"` // Идентификатор СМЭВ-сообщения от ведомства, сменившего статус
	RoutingCode   string `json:"routingCode"`   // Код маршрутизации СМЭВ-сообщения в ведомство

	PaymentRequired     bool             `json:"paymentRequired"`     // Флаг наличия оплаты
	NoPaidPaymentCount  int              `json:"noPaidPaymentCount"`  // Количество неоплаченных платежей
	PaymentCount        int              `json:"paymentCount"`        // Количество платежей
	HasNoPaidPayment    bool             `json:"hasNoPaidPayment"`    // Флаг наличия оплаченного платежа
	PaymentStatusEvents []map[string]any `json:"paymentStatusEvents"` // Статус событий при оплате [структура события оплаты не документирована]
	OrderPayments       []map[string]any `json:"orderPayments"`       // Информация о платежах [структура объекта платежа не документирована]
	Payback             bool             `json:"payback"`             // Служебный параметр

	ReadyToSign    bool   `json:"readyToSign"`    // Для ЮЛ, для подписания заявки, маркер ожидания УКЭП
	SignCnt        int    `json:"signCnt"`        // Кол-во подписей, для заявлений от нескольких заявителей
	AllFileSign    bool   `json:"allFileSign"`    // Флаг наличия ЭП для файлов
	ChildrenSigned bool   `json:"childrenSigned"` // Флаг подписи дочерних заявлений
	EdsStatus      string `json:"edsStatus"`      // Идентификатор статуса проверки ЭП

	TextMessages     []map[string]any `json:"textMessages"`     // Информация о сообщениях [структура объекта текстового сообщения не документирована]
	InfoMessages     []map[string]any `json:"infoMessages"`     // Информация о сообщениях [структура объекта сообщения не документирована]
	UnreadMessageCnt int              `json:"unreadMessageCnt"` // Кол-во непрочитанных сообщений

	NotifySms   string `json:"notifySms"`   // Флаг необходимости уведомления о смене статуса через СМС
	NotifyEmail string `json:"notifyEmail"` // Флаг необходимости уведомления о смене статуса через сообщение на эл. почту
	NotifyPush  string `json:"notifyPush"`  // Флаг необходимости уведомления о смене статуса через push-сообщение

	Qrlink OrderQrlink `json:"qrlink"` // [Не документировано]
	Steps  []any       `json:"steps"`  // [Не документировано]

}

// OrderStatus - статусы заявления структуры [Order]
type OrderStatus struct {

	// Основные аттрибуты

	Id                  int       `json:"id"`                  // Идентификатор статуса
	StatusId            int       `json:"statusId"`            // Код статуса
	Title               string    `json:"title"`               // Наименование статуса
	Date                time.Time `json:"date"`                // Дата и время смены статуса todo Формат!
	OrderId             int       `json:"orderId"`             // Номер заявления
	FinalStatus         bool      `json:"finalStatus"`         // Флаг финального статуса
	HasResult           string    `json:"hasResult"`           // Флаг передачи файла в ответ на заявление
	CancelAllowed       bool      `json:"cancelAllowed"`       // Флаг возможности отменить заявление
	Sender              string    `json:"sender"`              // Отправитель СМЭВ-сообщения о смене статуса
	Comment             string    `json:"comment"`             // Комментарий к статусу
	StateOrgStatusCode  string    `json:"stateOrgStatusCode"`  // Код ведомственного статуса
	StateOrgStatusDescr string    `json:"stateOrgStatusDescr"` // Наименование ведомственного статуса

	// Дополнительные аттрибуты

	UnreadEvent           bool   `json:"unreadEvent"`           // Признак прочтения
	DeliveryCancelAllowed bool   `json:"deliveryCancelAllowed"` // Флаг наличия отмены доставки
	SendMessageAllowed    bool   `json:"sendMessageAllowed"`    // Признак разрешения отправки сообщения
	EditAllowed           bool   `json:"editAllowed"`           // Признак редактирования
	Mnemonic              string `json:"mnemonic"`              // Мнемоника ИС отправителя
	StatusColorCode       string `json:"statusColorCode"`       // [Не документировано]

}

// OrderStatusHistory - история статуса заявления структуры [Order].
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
type OrderStatusHistory struct {

	// Основные аттрибуты

	Id                  int    `json:"id"`                  // Идентификатор статуса
	StatusId            int    `json:"statusId"`            // Код статуса
	Title               string `json:"title"`               // Наименование статуса
	Date                string `json:"date"`                // Дата и время смены статуса todo FORMAT!
	OrderId             int    `json:"orderId"`             // Номер заявления
	FinalStatus         bool   `json:"finalStatus"`         // Флаг финального статуса
	HasResult           string `json:"hasResult"`           // Флаг передачи файла в ответ на заявление
	CancelAllowed       bool   `json:"cancelAllowed"`       // Флаг наличия отмены
	Sender              string `json:"sender"`              // Наименование ведомства
	Comment             string `json:"comment"`             // Комментарий
	StateOrgStatusCode  string `json:"stateOrgStatusCode"`  // Код ведомственного статуса
	StateOrgStatusDescr string `json:"stateOrgStatusDescr"` // Наименование ведомственного статуса
	StatusColorCode     string `json:"statusColorCode"`     // [Не документировано]

	// Дополнительные аттрибуты

	UnreadEvent           bool `json:"unreadEvent"`           // Признак прочтения события
	DeliveryCancelAllowed bool `json:"deliveryCancelAllowed"` // Флаг наличия отмены доставки
	SendMessageAllowed    bool `json:"sendMessageAllowed"`    // Признак разрешения отправки сообщения
	EditAllowed           bool `json:"editAllowed"`           // Признак редактирования

}

// OrderAttachmentFile - файл заявления, отправленный пользователем из структуры [Order].
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
type OrderAttachmentFile struct {
	// Основные аттрибуты

	Id                  string `json:"id"`                  // Идентификатор файла
	FileName            string `json:"fileName"`            // Название файла
	MimeType            string `json:"mimeType"`            // MIME-тип
	Link                string `json:"link"`                // Ссылка на файл в хранилище
	HasDigitalSignature bool   `json:"hasDigitalSignature"` // Наличие подписи

	// Дополнительные аттрибуты

	FileSize int    `json:"fileSize"` // Размер файла
	Type     string `json:"type"`     // Тип
}

// OrderResponseFile - информация о файле в ответе заявления из структуры [Order].
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
type OrderResponseFile struct {
	// Основные аттрибуты

	Id                  string `json:"id"`                  // Идентификатор файла
	FileName            string `json:"fileName"`            // Наименование файла
	MimeType            string `json:"mimeType"`            // MIME-тип файла
	Link                string `json:"link"`                // Ссылка на файл в TERRABYTE
	HasDigitalSignature bool   `json:"hasDigitalSignature"` // Флаг наличия ЭП к файлу

	// Дополнительные аттрибуты

	HasAltMimeType bool   `json:"hasAltMimeType"` // Флаг наличия альтернативного MIME-типа
	EdsStatus      string `json:"edsStatus"`      // Статус проверки ЭП в EDS
	FileSize       int    `json:"fileSize"`       // Размер файла
}

// OrderAttributeEvent - атрибуты событий для заявления из структуры [Order].
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
type OrderAttributeEvent struct {
	Name     string `json:"name"`     // Наименование атрибута
	NewValue string `json:"newValue"` // Новое значение
	OldValue string `json:"oldValue"` // Старое значение
}

// OrderQrlink - не документированное поле (см [Order]).
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12", раздел "2.4. Получение деталей по заявлению".
type OrderQrlink struct {
	HasAltMimeType      bool `json:"hasAltMimeType"`      // Связанно с alternativeMimeTypes из сервиса тербайта
	FileSize            int  `json:"fileSize"`            // Размер файла
	HasDigitalSignature bool `json:"hasDigitalSignature"` // Флаг наличия ЭП
	CanSentToMFC        bool `json:"canSentToMFC"`        // Флаг отправки в МФЦ
	CanPrintMFC         bool `json:"canPrintMFC"`         // [Не документировано]
}
