package apipgu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/ofstudio/go-api-epgu/dto"
	"github.com/ofstudio/go-api-epgu/utils"
)

// DefaultChunkSize - размер чанка по умолчанию для метода [Client.OrderPushChunked].
// Если размер архива вложения будет больше, то метод отправит архив несколькими запросами.
// Значение можно изменить с помощью [Client.WithChunkSize].
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.3 Отправка заявления (загрузка архива по частям)".
const DefaultChunkSize = 5_000_000

// DefaultArchiveName - имя архива по умолчанию для методов [Client.OrderPush] и [Client.OrderPushChunked].
// Используется, если в [Archive].Name не передано имя архива.
const DefaultArchiveName = "archive"

// Client - REST-клиент для API Госуслуг.
type Client struct {
	baseURI    string
	httpClient *http.Client
	chunkSize  int
	debug      bool
	logger     utils.Logger
}

// NewClient - конструктор [Client].
func NewClient(baseURI string) *Client {
	return &Client{
		baseURI:    baseURI,
		httpClient: &http.Client{},
		chunkSize:  DefaultChunkSize,
	}
}

// WithDebug - включает логирование HTTP-запросов и ответов к ЕПГУ.
// Формат лога:
//
//	>>> Request to {url}
//	...
//	{полный HTTP-запрос}
//	...
//	<<< Response from {url}
//	...
//	{полный HTTP-ответ}
//	...
func (c *Client) WithDebug(logger utils.Logger) *Client {
	c.logger = logger
	c.debug = logger != nil
	return c
}

// WithHTTPClient - устанавливает http-клиент для запросов к ЕПГУ.
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	if httpClient != nil {
		c.httpClient = httpClient
	}
	return c
}

// WithChunkSize устанавливает максимальный размер чанка для метода [Client.OrderPushChunked].
// По умолчанию используется [DefaultChunkSize].
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.3 Отправка заявления (загрузка архива по частям)"
func (c *Client) WithChunkSize(n int) *Client {
	if n > 0 {
		c.chunkSize = n
	}
	return c
}

// OrderCreate - создание заявления.
//
//	POST /api/gusmev/order
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.2 Создание заявления".
//
// В случае успеха возвращает номер созданного заявления.
// В случае ошибки возвращает цепочку из [ErrOrderCreate] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - [ErrWrongOrderID] - в ответе не передан ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderCreate(token string, meta OrderMeta) (int, error) {
	orderIdResponse := &dto.OrderIdResponse{}
	if err := c.requestJSON(
		http.MethodPost,
		"/api/gusmev/order",
		"application/json; charset=utf-8",
		token,
		bytes.NewReader(meta.JSON()),
		orderIdResponse,
	); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrOrderCreate, err)
	}
	if orderIdResponse.OrderId == 0 {
		return 0, fmt.Errorf("%w: %w", ErrOrderCreate, ErrWrongOrderID)
	}
	return orderIdResponse.OrderId, nil
}

// OrderPushChunked - загрузка архива по частям.
//
//	POST /api/gusmev/push/chunked
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.3 Отправка заявления (загрузка архива по частям)"
//
// Максимальный размер чанка по умолчанию: [DefaultChunkSize],
// может быть изменен с помощью [Client.WithChunkSize].
//
// В случае ошибки возвращает цепочку из [ErrPushChunked] и следующих возможных ошибок:
//   - [ErrNilArchive] - не передан архив
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrMultipartBody] - ошибка подготовки multipart-содержимого
//   - [ErrWrongOrderID] - в ответе не передан или передан некорректный ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPushChunked(token string, orderId int, meta OrderMeta, archive *Archive) error {
	if archive == nil || len(archive.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrPushChunked, ErrNilArchive)
	}

	filename := archive.Name
	if archive.Name == "" {
		filename = DefaultArchiveName
	}
	extension := ".zip"

	total := 1 + (len(archive.Data)-1)/(c.chunkSize)

	for current := 0; current < total; current++ {
		// prepare chunk
		end := current*c.chunkSize + c.chunkSize
		if end > len(archive.Data) {
			end = len(archive.Data)
		}
		chunk := archive.Data[current*c.chunkSize : end]

		if total > 1 {
			extension = fmt.Sprintf(".z%03d", current+1)
		}

		// prepare multipart body
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		if err := newMultipartBuilder(w).
			withOrderId(orderId).
			withMeta(meta).
			withFile(filename+extension, chunk).
			withChunkNum(current, total).
			build(); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
		}

		// make request
		orderIdResponse := &dto.OrderIdResponse{}
		if err := c.requestJSON(
			http.MethodPost,
			"/api/gusmev/push/chunked",
			"multipart/form-data; boundary="+w.Boundary(),
			token,
			body,
			orderIdResponse,
		); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
		}
		if orderIdResponse.OrderId != orderId {
			return fmt.Errorf("%w: %w", ErrPushChunked, ErrWrongOrderID)
		}
	}

	return nil
}

// OrderPush - формирование заявления единым методом.
//
//	POST /api/gusmev/push
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.4 Формирование заявления единым методом"
//
// В случае успеха возвращает номер созданного заявления.
// В случае ошибки возвращает цепочку из [ErrPush] и следующих возможных ошибок:
//   - [ErrNilArchive] - не передан архив
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrMultipartBody] - ошибка подготовки multipart-содержимого
//   - [ErrWrongOrderID] - в ответе не передан ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPush(token string, meta OrderMeta, archive *Archive) (int, error) {
	if archive == nil || len(archive.Data) == 0 {
		return 0, fmt.Errorf("%w: %w", ErrPush, ErrNilArchive)
	}

	filename := archive.Name
	if archive.Name == "" {
		filename = DefaultArchiveName
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := newMultipartBuilder(w).
		withMeta(meta).
		withFile(filename+".zip", archive.Data).
		build(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPush, err)
	}

	orderIdResponse := &dto.OrderIdResponse{}
	if err := c.requestJSON(
		http.MethodPost,
		"/api/gusmev/push",
		"multipart/form-data; boundary="+w.Boundary(),
		token,
		body,
		orderIdResponse,
	); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPush, err)
	}
	if orderIdResponse.OrderId == 0 {
		return 0, fmt.Errorf("%w: %w", ErrPush, ErrWrongOrderID)
	}

	return orderIdResponse.OrderId, nil
}

// OrderInfo - запрос детальной информации по отправленному заявлению.
//
//	POST /api/gusmev/order/{orderId}
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.4. Получение деталей по заявлению".
//
// В случае успеха возвращает детальную информацию по заявлению.
// В случае ошибки возвращает цепочку из [ErrOrderInfo] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderInfo(token string, orderId int) (*OrderInfo, error) {

	orderInfoResponse := &dto.OrderInfoResponse{}
	if err := c.requestJSON(
		http.MethodPost,
		fmt.Sprintf("/api/gusmev/order/%d", orderId),
		"",
		token,
		nil,
		orderInfoResponse,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOrderInfo, err)
	}

	orderInfo := &OrderInfo{
		Code:      orderInfoResponse.Code,
		Message:   orderInfoResponse.Message,
		MessageId: orderInfoResponse.MessageId,
	}

	// unmarshal order field
	if orderInfoResponse.Order != "" {
		orderInfo.Order = &OrderDetails{}
		if err := json.Unmarshal([]byte(orderInfoResponse.Order), orderInfo.Order); err != nil {
			return nil, fmt.Errorf("%w: %w: %w", ErrOrderInfo, ErrJSONUnmarshal, err)
		}
	}

	return orderInfo, nil
}

// OrderCancel - отмена заявления.
//
//	POST /api/gusmev/order/{orderId}/cancel
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.2. Отмена заявления".
//
// В случае ошибки возвращает цепочку из [ErrOrderCancel] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeCancelNotAllowed])
//
// Примечание. В настоящий момент (декабрь 2023) вызов метода возвращает ошибку HTTP 400 Bad Request:
//
//	 {
//		 "code":"bad_request",
//		 "message":"Required request parameter 'reason' for method parameter type String is not present"
//	 }
//
// При этом, параметр reason не описан в спецификации.
// На данный момент ни одна из доступных услуг API ЕПГУ не предусматривает
// возможность отмены. Вероятно, спецификация метода будет изменена в будущем.
func (c *Client) OrderCancel(token string, orderId int) error {
	if _, err := c.requestBody(
		http.MethodPost,
		fmt.Sprintf("/api/gusmev/order/%d/cancel", orderId),
		"application/json; charset=utf-8",
		token,
		nil,
	); err != nil {
		return fmt.Errorf("%w: %w", ErrOrderCancel, err)
	}
	return nil
}

// AttachmentDownload - скачивание файла вложения созданного заявления.
//
//	GET /api/storage/v2/files/{objectId}/{objectType}/download?mnemonic={mnemonic}
//
// Параметр link - значение поля [OrderAttachmentFile].Link из ответа метода [Client.OrderInfo].
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "4. Скачивание файла".
//
// В случае успеха возвращает содержимое файла.
// В случае ошибки возвращает цепочку из [ErrAttachmentDownload] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrInvalidFileLink] - некорректный параметр link
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeAccessDeniedSystem])
func (c *Client) AttachmentDownload(token string, link string) ([]byte, error) {
	uri, err := attachmentURI(link)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAttachmentDownload, err)
	}

	resBody, err := c.requestBody(
		http.MethodGet,
		"/api/storage/v2/files"+uri,
		"",
		token,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAttachmentDownload, err)
	}

	return resBody, nil
}

// reAttachmentURI - регулярное выражение для разбора URI вида:
// "terrabyte://00/1230254874/req_8d8567db-d445-4759-a122-6b4cefeca22c.xml/2"
// $1 - {objectId}
// $2 - {mnemonic}
// $3 - {objectType}
var reAttachmentURI = regexp.MustCompile(`^terrabyte://.*/(.*)/(.*)/(.*)$`)

// attachmentURI - формирует URI для скачивания файла вложения.
// Параметр link - значение поля [OrderAttachmentFile.Link].
// Возвращает URI вида:
//
//	/{objectId}/{objectType}/download?mnemonic={mnemonic}
//
// либо ошибку [ErrInvalidFileLink], если передан некорректный параметр link.
func attachmentURI(link string) (string, error) {
	matches := reAttachmentURI.FindStringSubmatch(link)
	if len(matches) != 4 {
		return "", ErrInvalidFileLink
	}
	return fmt.Sprintf("/%s/%s/download?mnemonic=%s", matches[1], matches[3], matches[2]), nil
}

// Dict - получение справочных данных.
//
//	POST /api/nsi/v1/dictionary/{code}
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "3. Получение справочных данных".
//
// Параметры:
//
//   - code - код справочника. Примеры: "EXTERNAL_BIC", "TO_PFR"
//   - filter - тип справочника (плоский [DictFilterOneLevel] или иерархический [DictFilterSubTree])
//   - parent - код родительского элемента (необязательный)
//   - pageNum - номер необходимой страницы (необязательный)
//   - pageSize - количество записей на странице (необязательный)
//
// Примечание: не все справочники поддерживают параметры parent, pageNum и pageSize.
func (c *Client) Dict(code string, filter, parent string, pageNum, pageSize int) (*Dict, error) {
	reqBody, _ := json.Marshal(&dto.DictRequest{
		TreeFiltering:      filter,
		ParentRefItemValue: parent,
		PageNum:            pageNum,
		PageSize:           pageSize,
	})

	dict := &Dict{}
	if err := c.requestJSON(
		http.MethodPost,
		fmt.Sprintf("/api/nsi/v1/dictionary/%s", code),
		"application/json; charset=utf-8",
		"",
		bytes.NewReader(reqBody),
		dict,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDict, err)
	}

	return dict, nil
}
