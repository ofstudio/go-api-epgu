package apipgu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

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
//   - [ErrNoOrderID] - не передан ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderCreate(token string, meta OrderMeta) (int, error) {
	orderIdResponse := &dto.OrderIdResponse{}
	if err := c.request(
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
		return 0, fmt.Errorf("%w: %w", ErrOrderCreate, ErrNoOrderID)
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
// В случае ошибки возвращает цепочку из [ErrPushChunked] и следующих возможных ошибок:
//   - [ErrNilArchive] - не передан архив
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrMultipartBody] - ошибка подготовки multipart-содержимого
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPushChunked(token string, id int, meta OrderMeta, archive *Archive) error {
	if archive == nil || len(archive.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrPushChunked, ErrNilArchive)
	}

	total := 1 + len(archive.Data)/(c.chunkSize+1)

	for current := 0; current < total; current++ {
		// prepare chunk
		end := current*c.chunkSize + c.chunkSize
		if end > len(archive.Data) {
			end = len(archive.Data)
		}
		chunk := archive.Data[current*c.chunkSize : end]

		filename := archive.Name
		if total > 1 {
			filename = archive.Name + fmt.Sprintf(".z%03d", current+1)
		} else {
			filename += ".zip"
		}

		// prepare multipart body
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		if err := newMultipartBuilder(w).
			withOrderId(id).
			withMeta(meta).
			withFile(filename, chunk).
			withChunkNum(current, total).
			build(); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
		}

		// make request
		orderIdResponse := &dto.OrderIdResponse{}
		if err := c.request(
			http.MethodPost,
			"/api/gusmev/push/chunked",
			"multipart/form-data; boundary="+w.Boundary(),
			token,
			body,
			orderIdResponse,
		); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
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
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPush(token string, meta OrderMeta, archive *Archive) (int, error) {
	if archive == nil || len(archive.Data) == 0 {
		return 0, fmt.Errorf("%w: %w", ErrPush, ErrNilArchive)
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := newMultipartBuilder(w).
		withMeta(meta).
		withFile(archive.Name+".zip", archive.Data).
		build(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPush, err)
	}

	orderIdResponse := &dto.OrderIdResponse{}
	if err := c.request(
		http.MethodPost,
		"/api/gusmev/push",
		"multipart/form-data; boundary="+w.Boundary(),
		token,
		body,
		orderIdResponse,
	); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPush, err)
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
// В случае ошибки возвращает цепочку из ErrOrderInfo и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderInfo(token string, orderId int) (*OrderInfo, error) {

	orderInfoResponse := &dto.OrderInfoResponse{}
	if err := c.request(
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
