package apipgu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/ofstudio/go-api-epgu/utils"
)

// DefaultChunkSize - размер чанка по умолчанию для метода [Client.OrderPushChunked].
// Если размер архива вложения будет больше, то метод отправит архив несколькими запросами.
// Значение можно изменить с помощью [Client.WithChunkSize].
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.14",
// раздел "2.1.3 Загрузка и отправка заявления после резервирования номера".
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
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.1.3 Загрузка и отправка заявления после резервирования номера"
func (c *Client) WithChunkSize(n int) *Client {
	if n > 0 {
		c.chunkSize = n
	}
	return c
}

// OrderCreate - резервирование номера заявления.
//
//	POST /api/gusmev/order
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.14",
// раздел "2.1.2 Резервирование номера заявления".
//
// В случае успеха возвращает номер созданного заявления.
// В случае ошибки возвращает цепочку из [ErrOrderCreate] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - [ErrWrongOrderID] - в ответе не передан ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderCreate(ctx context.Context, token string, meta OrderMeta) (int, error) {
	orderIdResponse := &dtoOrderIdResponse{}
	if err := c.requestJSON(
		ctx,
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

// OrderPushChunked - загрузка и отправка заявления после резервирования номера.
//
//	POST /api/gusmev/push/chunked
//
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.1.3 Загрузка и отправка заявления после резервирования номера"
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
//
// Примечание: согласно спецификации поле meta является обязательным
// для запроса /api/gusmev/push/chunked. Но это, вероятнее всего,
// опечатка в спецификации. Данный метод не отправляет поле meta.
func (c *Client) OrderPushChunked(ctx context.Context, token string, orderId int, archive *Archive) error {
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
		builder := newMultipartBuilder(w).
			withOrderId(orderId).
			withFile(filename+extension, chunk)
		if total > 1 {
			builder = builder.withChunkNum(current, total)
		}
		if err := builder.build(); err != nil {
			return fmt.Errorf("%w: %w", ErrPushChunked, err)
		}

		// make request
		orderIdResponse := &dtoOrderIdResponse{}
		if err := c.requestJSON(
			ctx,
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

// OrderPush - загрузка и отправка заявления одним запросом.
//
//	POST /api/gusmev/push
//
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.1.4 Загрузка и отправка заявления одним запросом"
//
// В случае успеха возвращает номер созданного заявления.
// В случае ошибки возвращает цепочку из [ErrPush] и следующих возможных ошибок:
//   - [ErrNilArchive] - не передан архив
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrMultipartBody] - ошибка подготовки multipart-содержимого
//   - [ErrWrongOrderID] - в ответе не передан ID заявления
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderPush(ctx context.Context, token string, meta OrderMeta, archive *Archive) (int, error) {
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

	orderIdResponse := &dtoOrderIdResponse{}
	if err := c.requestJSON(
		ctx,
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
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.4. Получение деталей по заявлению".
//
// В случае успеха возвращает детальную информацию по заявлению.
// В случае ошибки возвращает цепочку из [ErrOrderInfo] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) OrderInfo(ctx context.Context, token string, orderId int) (*OrderInfo, error) {

	orderInfoResponse := &dtoOrderInfoResponse{}
	if err := c.requestJSON(
		ctx,
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
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.2. Отмена заявления".
//
// В случае ошибки возвращает цепочку из [ErrOrderCancel] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeCancelNotAllowed])
//
// На данный момент ни одна из доступных услуг API ЕПГУ не предусматривает
// возможность отмены. Вероятно, спецификация метода будет изменена в будущем.
func (c *Client) OrderCancel(ctx context.Context, token string, orderId int) error {
	if _, err := c.requestBody(
		ctx,
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
//	GET /api/gusmev/files/download/{objectId}/{objectType}?mnemonic={mnemonic}&eserviceCode={eserviceCode}
//
// Параметр currentStatusHistoryId - значение поля [OrderDetails].CurrentStatusHistoryId.
// Параметр link - значение поля [OrderAttachmentFile].Link или [OrderResponseFile].Link
// из ответа метода [Client.OrderInfo].
// Параметр eserviceCode - код услуги.
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.6. Скачивание файла".
//
// В случае успеха возвращает содержимое файла.
// В случае ошибки возвращает цепочку из [ErrAttachmentDownload] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrInvalidFileLink] - некорректный параметр link
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeAccessDeniedSystem])
func (c *Client) AttachmentDownload(ctx context.Context, token string, currentStatusHistoryId int, link, eserviceCode string) ([]byte, error) {
	body := &bytes.Buffer{}
	if err := c.AttachmentDownloadTo(ctx, token, currentStatusHistoryId, link, eserviceCode, body); err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

// AttachmentDownloadTo - скачивание файла вложения созданного заявления в поток.
//
//	GET /api/gusmev/files/download/{objectId}/{objectType}?mnemonic={mnemonic}&eserviceCode={eserviceCode}
//
// Параметр currentStatusHistoryId - значение поля [OrderDetails].CurrentStatusHistoryId.
// Параметр link - значение поля [OrderAttachmentFile].Link или [OrderResponseFile].Link
// из ответа метода [Client.OrderInfo].
// Параметр eserviceCode - код услуги.
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.6. Скачивание файла".
//
// В случае ошибки возвращает цепочку из [ErrAttachmentDownload] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrInvalidFileLink] - некорректный параметр link
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusUnauthorized])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeAccessDeniedSystem])
func (c *Client) AttachmentDownloadTo(ctx context.Context, token string, currentStatusHistoryId int, link, eserviceCode string, dst io.Writer) error {
	endpoint, err := attachmentEndpoint(currentStatusHistoryId, link, eserviceCode)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAttachmentDownload, err)
	}

	if err = c.requestStream(ctx, http.MethodGet, endpoint, "", token, nil, dst); err != nil {
		return fmt.Errorf("%w: %w", ErrAttachmentDownload, err)
	}

	return nil
}

// parseAttachmentLink - разбирает URI вида:
// "terrabyte://00/1230254874/req_8d8567db-d445-4759-a122-6b4cefeca22c.xml/2"
// и возвращает mnemonic и objectType.
func parseAttachmentLink(link string) (mnemonic, objectType string, err error) {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "terrabyte" {
		return "", "", ErrInvalidFileLink
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", ErrInvalidFileLink
	}

	mnemonic = parts[len(parts)-2]
	objectType = parts[len(parts)-1]
	if mnemonic == "" || objectType == "" {
		return "", "", ErrInvalidFileLink
	}
	return mnemonic, objectType, nil
}

// attachmentEndpoint - формирует endpoint для скачивания файла вложения.
// Параметр currentStatusHistoryId - значение поля [OrderDetails.CurrentStatusHistoryId].
// Параметр link - значение поля [OrderAttachmentFile.Link] или [OrderResponseFile.Link].
// Возвращает endpoint вида:
//
//	/api/gusmev/files/download/{objectId}/{objectType}?mnemonic={mnemonic}&eserviceCode={eserviceCode}
//
// либо ошибку [ErrInvalidFileLink], если передан некорректный параметр link.
func attachmentEndpoint(currentStatusHistoryId int, link, eserviceCode string) (string, error) {
	if currentStatusHistoryId <= 0 {
		return "", ErrInvalidFileLink
	}

	mnemonic, objectType, err := parseAttachmentLink(link)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("mnemonic", mnemonic)
	query.Set("eserviceCode", eserviceCode)

	return fmt.Sprintf(
		"/api/gusmev/files/download/%s/%s?%s",
		url.PathEscape(fmt.Sprintf("%d", currentStatusHistoryId)),
		url.PathEscape(objectType),
		query.Encode(),
	), nil
}

// Dict - получение справочных данных.
//
//	POST /api/nsi/v1/dictionary/{code}
//
// Подробнее см "Спецификация API ЕПГУ версия 1.14",
// раздел "2.5. Получение справочных данных".
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
//
// В случае успеха возвращает элементы справочника с учетом pageNum и pageSize,
// а также общее количество найденных элементов.
// В случае ошибки возвращает цепочку из [ErrDict] и следующих возможных ошибок:
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - [ErrDictResponse] - ошибка получения справочных данных c указанием code и message из ответа
//   - HTTP-ошибок ErrStatusXXXX (например, [ErrStatusBadRequest])
//   - Ошибок ЕПГУ: ErrCodeXXXX (например, [ErrCodeBadRequest])
func (c *Client) Dict(ctx context.Context, code string, filter, parent string, pageNum, pageSize int) ([]DictItem, int, error) {
	reqBody, _ := json.Marshal(&dtoDictRequest{
		TreeFiltering:      filter,
		ParentRefItemValue: parent,
		PageNum:            pageNum,
		PageSize:           pageSize,
	})

	dictResponse := &dtoDictResponse{}
	if err := c.requestJSON(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/api/nsi/v1/dictionary/%s", code),
		"application/json; charset=utf-8",
		"",
		bytes.NewReader(reqBody),
		dictResponse,
	); err != nil {
		return nil, 0, fmt.Errorf("%w: %w", ErrDict, err)
	}

	if dictResponse.Error.Code != 0 && len(dictResponse.Items) == 0 {
		return nil, 0, fmt.Errorf("%w: %w", ErrDict, dictError(dictResponse.Error))
	}

	return dictResponse.Items, dictResponse.Total, nil
}
