package apipgu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
func (c *Client) WithDebug(logger utils.Logger) *Client {
	if c == nil {
		return nil
	}
	c.logger = logger
	c.debug = logger != nil
	return c
}

// WithHTTPClient - устанавливает http-клиент для запросов к ЕПГУ.
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	if c != nil && httpClient != nil {
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
	if c != nil && n > 0 {
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
//   - [ErrRequestPrepare], [ErrRequestCall], [ErrResponseRead] - ошибки выполнения запроса
//   - [ErrUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderCreate(token string, meta OrderMeta) (int, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURI+"/api/gusmev/order", bytes.NewReader(meta.JSON()))
	if err != nil {
		return 0, fmt.Errorf("%w: %w: %w", ErrOrderCreate, ErrRequestPrepare, err)
	}
	req.Header.Set("Content-Type", "application/JSON; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	c.logReq(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %w: %w", ErrOrderCreate, ErrRequestCall, err)
	}

	c.logRes(res)

	if res.StatusCode >= 400 {
		return 0, fmt.Errorf("%w: %w", ErrOrderCreate, responseError(res))
	}

	var orderResponse dto.OrderIdResponse
	//goland:noinspection ALL
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("%w: %w: %w", ErrOrderCreate, ErrResponseRead, err)
	}
	if err = json.Unmarshal(body, &orderResponse); err != nil {
		return 0, fmt.Errorf("%w: %w: %w", ErrOrderCreate, ErrUnmarshal, err)
	}

	return orderResponse.OrderId, nil
}

// OrderPushChunked - загрузка архива по частям
//
//	POST /api/gusmev/push/chunked
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.1.3 Отправка заявления (загрузка архива по частям)"
//
// В случае ошибки возвращает цепочку из [ErrPushChunked] и следующих возможных ошибок:
//   - [ErrRequestPrepare], [ErrRequestCall], [ErrResponseRead] - ошибки выполнения запроса
//   - [ErrMultipartBodyPrepare] - ошибка подготовки multipart-содержимого
//   - [ErrZipCreate], [ErrZipWrite], [ErrZipClose] - ошибки формирования zip-архива
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPushChunked(token string, id int, meta OrderMeta, archive PushArchive) error {
	zip, err := archive.Zip()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPushChunked, err)
	}
	total := 1 + len(zip)/(c.chunkSize+1)

	for current := 0; current < total; current++ {
		// prepare chunk
		end := current*c.chunkSize + c.chunkSize
		if end > len(zip) {
			end = len(zip)
		}
		chunk := zip[current*c.chunkSize : end]

		filename := archive.Name
		if total > 1 {
			filename = archive.Name + fmt.Sprintf(".z%03d", current+1)
		} else {
			filename += ".zip"
		}

		// prepare multipart body
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		if err = multipartBodyPrepare(
			w,
			withOrderId(id),
			withMeta(meta),
			withFile(filename, chunk),
			withChunkNum(current, total),
		); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
		}

		// make request
		req, err := http.NewRequest(http.MethodPost, c.baseURI+"/api/gusmev/push/chunked", body)
		if err != nil {
			return fmt.Errorf("%w: %w: %w", ErrPushChunked, ErrRequestPrepare, err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary="+w.Boundary())
		req.Header.Set("Authorization", "Bearer "+token)

		c.logReq(req)

		res, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("%w: %w: %w", ErrPushChunked, ErrRequestCall, err)
		}

		c.logRes(res)

		// todo 204 code
		if res.StatusCode >= 400 {
			return fmt.Errorf("%w: %w", ErrPushChunked, responseError(res))
		}
	}

	return nil
}

// OrderInfo - запрос детальной информации по отправленному заявлению.
//
//	POST /api/gusmev/order/{orderId}
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.4. Получение деталей по заявлению".
//
// В случае успеха возвращает детальную информацию по заявлению.
// В случае ошибки возвращает цепочку из ErrOrderInfo и  и следующих возможных ошибок:
//   - [ErrRequestPrepare], [ErrRequestCall], [ErrResponseRead] - ошибки выполнения запроса
//   - [ErrUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderInfo(token string, orderId int) (*dto.OrderInfoResponse, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/gusmev/order/%d", c.baseURI, orderId), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrOrderInfo, ErrRequestPrepare, err)
	}
	req.Header.Set("Content-Type", "application/JSON; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	c.logReq(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrOrderInfo, ErrRequestCall, err)
	}

	c.logRes(res)

	// todo 204 code
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: %w", ErrOrderInfo, responseError(res))
	}

	orderInfoRes := &dto.OrderInfoResponse{}
	//goland:noinspection ALL
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrOrderInfo, ErrResponseRead, err)
	}
	if err = json.Unmarshal(body, orderInfoRes); err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrOrderInfo, ErrUnmarshal, err)
	}

	return orderInfoRes, nil

}
