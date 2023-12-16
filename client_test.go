package apipgu

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(suiteTestClient))
}

type suiteTestClient struct {
	suite.Suite
}

func (suite *suiteTestClient) TestOrderCreate() {

	// Cases:
	//	1. success
	//	2. unexpected json response
	//	3. malformed json response
	//	4. bad request
	//	5. forbidden
	//	6. request error

	suite.Run("success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			suite.Equal("POST", r.Method)
			suite.Equal("/api/gusmev/order", r.URL.Path)
			suite.Equal("application/json; charset=utf-8", r.Header.Get("Content-Type"))
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))
			body, _ := io.ReadAll(r.Body)
			suite.JSONEq(`{"region":"test-region","serviceCode":"test-service","targetCode":"test-target"}`, string(body))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.NoError(err)
		suite.Equal(123456, orderId)
	})

	suite.Run("unexpected json response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"foo":"bar"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrWrongOrderID)
		suite.Equal(0, orderId)
	})

	suite.Run("malformed json response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`malformed json{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(0, orderId)
	})

	suite.Run("bad request", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Ошибка валидации запроса по схеме"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrStatusBadRequest)
		suite.Equal(
			"ошибка OrderCreate: HTTP 400 Bad Request: неверные параметры: код ошибки не указан [error='Ошибка валидации запроса по схеме']",
			err.Error(),
		)
		suite.Equal(0, orderId)
	})

	suite.Run("forbidden", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code":"access_denied_service", "message":"Доступ ВИС к запрашиваемой услуге запрещен"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrStatusForbidden)
		suite.Equal(
			"ошибка OrderCreate: HTTP 403 Forbidden: доступ запрещен: доступ ВИС к запрашиваемой услуге запрещен [code='access_denied_service', message='Доступ ВИС к запрашиваемой услуге запрещен']",
			err.Error(),
		)
		suite.Equal(0, orderId)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrRequest)
		suite.Equal(0, orderId)
	})
}

func (suite *suiteTestClient) TestOrderPushChunked() {

	// Cases:
	//	1. success single chunk less than chunkSize
	//	2. success single chunk equal to chunkSize
	//	3. success multiple chunks
	//	4. success but no orderId in response
	//	5. success but wrong orderId in response
	//	6. internal error with plain text response
	//	7. request error
	//  8. empty archive name
	//  9. archive is nil
	//  10. archive is zero length

	suite.Run("success single chunk less than chunkSize", func() {
		reqCount := 0
		dataSent := make([]byte, 100)
		_, err := rand.Read(dataSent)
		suite.Require().NoError(err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCount++
			suite.Equal("POST", r.Method)
			suite.Equal("/api/gusmev/push/chunked", r.URL.Path)
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))
			suite.Contains(r.Header.Get("Content-Type"), "multipart/form-data; boundary=")

			suite.NoError(r.ParseMultipartForm(0))
			suite.Equal("0", r.FormValue("chunk"))
			suite.Equal("1", r.FormValue("chunks"))
			suite.Equal("123456", r.FormValue("orderId"))
			suite.JSONEq(`{"region":"test-region","serviceCode":"test-service","targetCode":"test-target"}`, r.FormValue("meta"))
			suite.Len(r.MultipartForm.File["file"], 1)
			fh := r.MultipartForm.File["file"][0]
			suite.Equal("test-archive.zip", fh.Filename)
			suite.Equal("application/octet-stream", fh.Header.Get("Content-Type"))
			f, err := fh.Open()
			suite.Require().NoError(err)
			//goland:noinspection ALL
			defer f.Close()
			data, err := io.ReadAll(f)
			suite.Require().NoError(err)
			suite.Equal(dataSent, data)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: dataSent}
		suite.NoError(client.OrderPushChunked(testToken, 123456, testMeta, testArchive))
		suite.Equal(1, reqCount)
	})

	suite.Run("success single chunk equal to chunkSize", func() {
		reqCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCount++
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.NoError(err)
		suite.Equal(1, reqCount)
	})

	suite.Run("success multiple chunks", func() {
		reqCount := 0
		var dataReceived []byte
		var chunkSizes []int
		dataSent := make([]byte, 301)
		_, err := rand.Read(dataSent)
		suite.Require().NoError(err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCount++
			suite.NoError(r.ParseMultipartForm(0))
			suite.Equal("123456", r.FormValue("orderId"))
			suite.JSONEq(`{"region":"test-region","serviceCode":"test-service","targetCode":"test-target"}`, r.FormValue("meta"))
			suite.Len(r.MultipartForm.File["file"], 1)
			fh := r.MultipartForm.File["file"][0]
			suite.Equal(fmt.Sprintf("test-archive.z%03d", reqCount), fh.Filename)
			suite.Equal("application/octet-stream", fh.Header.Get("Content-Type"))
			f, err := fh.Open()
			suite.Require().NoError(err)
			//goland:noinspection ALL
			defer f.Close()
			data, err := io.ReadAll(f)
			suite.Require().NoError(err)
			dataReceived = append(dataReceived, data...)
			chunkSizes = append(chunkSizes, len(data))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: dataSent}
		suite.NoError(client.OrderPushChunked(testToken, 123456, testMeta, testArchive))
		suite.NoError(err)
		suite.Equal(4, reqCount)
		suite.Equal(testArchive.Data, dataReceived)
		suite.Equal([]int{100, 100, 100, 1}, chunkSizes)
	})

	suite.Run("success but no orderId in response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"foo":"bar"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrWrongOrderID)
	})

	suite.Run("success but wrong orderId in response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":9999}`))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrWrongOrderID)
	})

	suite.Run("internal error with plain text response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		}))
		defer server.Close()

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrStatusInternalError)
		suite.ErrorIs(err, ErrUnexpectedContentType)
		suite.Equal(
			"ошибка OrderPushChunked: HTTP 500 Internal Server Error: внутренняя ошибка: неожиданный тип содержимого: 'text/plain; charset=utf-8'",
			err.Error(),
		)
	})

	suite.Run("request error", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrRequest)
	})

	suite.Run("empty archive name", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.NoError(r.ParseMultipartForm(0))
			suite.Len(r.MultipartForm.File["file"], 1)
			fh := r.MultipartForm.File["file"][0]
			suite.Equal("archive.zip", fh.Filename)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))

		client := NewClient(server.URL).WithChunkSize(100)
		testArchive := &Archive{Name: "", Data: bytes.Repeat([]byte("a"), 100)}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.NoError(err)
	})

	suite.Run("archive is nil", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: nil}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrNilArchive)
	})

	suite.Run("archive is zero length", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: []byte{}}
		err := client.OrderPushChunked(testToken, 123456, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrNilArchive)
	})

}

func (suite *suiteTestClient) TestOrderPush() {

	// Cases:
	//	1. success
	//	2. success but no orderId in response
	//	3. 409 unable to handle request with code = service_not_found
	//	4. 400 with malformed json response
	//	5. request error
	//  6. archive is nil
	//  7. archive is zero length
	//  8. empty archive name

	suite.Run("success", func() {
		dataSent := make([]byte, 100)
		_, err := rand.Read(dataSent)
		suite.Require().NoError(err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal("POST", r.Method)
			suite.Equal("/api/gusmev/push", r.URL.Path)
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))
			suite.Contains(r.Header.Get("Content-Type"), "multipart/form-data; boundary=")

			suite.NoError(r.ParseMultipartForm(0))
			suite.JSONEq(`{"region":"test-region","serviceCode":"test-service","targetCode":"test-target"}`, r.FormValue("meta"))
			suite.Len(r.MultipartForm.File["file"], 1)
			fh := r.MultipartForm.File["file"][0]
			suite.Equal("test-archive.zip", fh.Filename)
			suite.Equal("application/octet-stream", fh.Header.Get("Content-Type"))
			f, err := fh.Open()
			suite.Require().NoError(err)
			//goland:noinspection ALL
			defer f.Close()
			data, err := io.ReadAll(f)
			suite.Require().NoError(err)
			suite.Equal(dataSent, data)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		testArchive := &Archive{Name: "test-archive", Data: dataSent}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.NoError(err)
		suite.Equal(123456, orderId)
	})

	suite.Run("success but no orderId in response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"foo":"bar"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrWrongOrderID)
		suite.Equal(0, orderId)
	})

	suite.Run("409 unable to handle request with code = service_not_found", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(409)
			_, _ = w.Write([]byte(`{"code":"service_not_found", "message":"Услуга не найдена"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrStatusUnableToHandleRequest)
		suite.ErrorIs(err, ErrCodeServiceNotFound)
		suite.Equal(
			"ошибка OrderPush: HTTP 409 Conflict: невозможно обработать запрос: не найдена услуга, заданная кодом serviceCode в запросе [code='service_not_found', message='Услуга не найдена']",
			err.Error(),
		)
		suite.Equal(0, orderId)
	})

	suite.Run("400 with malformed json response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`malformed json{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrStatusBadRequest)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(
			"ошибка OrderPush: HTTP 400 Bad Request: неверные параметры: ошибка чтения JSON: invalid character 'm' looking for beginning of value",
			err.Error(),
		)
		suite.Equal(0, orderId)
	})

	suite.Run("request error", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: bytes.Repeat([]byte("a"), 100)}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrRequest)
		suite.Equal(0, orderId)
	})

	suite.Run("archive is nil", func() {
		client := NewClient("")
		orderId, err := client.OrderPush(testToken, testMeta, nil)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrNilArchive)
		suite.Equal(0, orderId)
	})

	suite.Run("archive is zero length", func() {
		client := NewClient("")
		orderId, err := client.OrderPush(testToken, testMeta, &Archive{})
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrNilArchive)
		suite.Equal(0, orderId)
	})

	suite.Run("empty archive name", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.NoError(r.ParseMultipartForm(0))
			suite.Len(r.MultipartForm.File["file"], 1)
			fh := r.MultipartForm.File["file"][0]
			suite.Equal("archive.zip", fh.Filename)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))

		client := NewClient(server.URL)
		testArchive := &Archive{Name: "", Data: bytes.Repeat([]byte("a"), 100)}
		orderId, err := client.OrderPush(testToken, testMeta, testArchive)
		suite.NoError(err)
		suite.Equal(123456, orderId)
	})
}

func (suite *suiteTestClient) TestOrderInfo() {

	// Cases:
	//	1. success
	//	2. success but order is nil
	//	3. malformed json response
	//	4. malformed order field
	//	5. 204 - order not found
	//	6. request error
	//	7. 401 - unauthorized
	//	8. unexpected 4xx status
	//	9. unknown code in error response

	suite.Run("success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/gusmev/order/123456", r.URL.Path)
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(orderInfoResponseJSON))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.NoError(err)
		suite.NotNil(orderInfo)
		suite.Equal("OK", orderInfo.Code)
		suite.Equal("test", orderInfo.Message)
		suite.Equal("test-GUID", orderInfo.MessageId)
		suite.NotNil(orderInfo.Order)
		orderJSON, err := json.Marshal(orderInfo.Order)
		suite.NoError(err)
		suite.JSONEq(orderInfoOrderWantJSON, string(orderJSON))
	})

	suite.Run("success but order is nil", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":"OK","message":"test","messageId":"test-GUID","order":null}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.NoError(err)
		suite.NotNil(orderInfo)
		suite.Equal("OK", orderInfo.Code)
		suite.Equal("test", orderInfo.Message)
		suite.Equal("test-GUID", orderInfo.MessageId)
		suite.Nil(orderInfo.Order)
	})

	suite.Run("malformed json response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`malformed json{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Nil(orderInfo)
	})

	suite.Run("malformed order field", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":"OK","message":"test","messageId":"test-GUID","order":"malformed json{}"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Nil(orderInfo)
	})

	suite.Run("204 - order not found", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusOrderNotFound)
		suite.Nil(orderInfo)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrRequest)
		suite.Nil(orderInfo)
	})

	suite.Run("401 - unauthorized", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusUnauthorized)
		suite.Nil(orderInfo)
	})

	suite.Run("unexpected 4xx status", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(406)
			_, _ = w.Write([]byte(`{"code":"access_denied_system", "message":"Access Denied"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusUnexpected)
		suite.Equal(
			"ошибка OrderInfo : HTTP 406 Not Acceptable: неожиданный HTTP-статус: доступ запрещен для ВИС, отправляющей запрос [code='access_denied_system', message='Access Denied']",
			err.Error(),
		)
		suite.Nil(orderInfo)
	})

	suite.Run("unknown code in error response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"code":"unknown_code", "message":"Unknown Code"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusBadRequest)
		suite.ErrorIs(err, ErrCodeUnexpected)
		suite.Equal(
			"ошибка OrderInfo : HTTP 400 Bad Request: неверные параметры: неожиданный код ошибки [code='unknown_code', message='Unknown Code']",
			err.Error(),
		)
		suite.Nil(orderInfo)
	})
}

var (
	testToken = "test-token"
	testMeta  = OrderMeta{Region: "test-region", ServiceCode: "test-service", TargetCode: "test-target"}
)

const orderInfoResponseJSON = `{"code":"OK","message":"test","messageId":"test-GUID","order":"{\"orderType\":\"ORDER\",\"hasNewStatus\":true,\"smevTx\":\"14329ae4-875b-40d3-9367-e2ef1b135189\",\"stateOrgId\":266,\"hasEmpowerment2021\":false,\"smevMessageId\":\"WAIT_RESPONSE\",\"hasChildren\":false,\"stateStructureName\":\"СФР\",\"formVersion\":\"1\",\"possibleServices\":{},\"orderAttributeEvents\":[],\"ownerId\":1000571421,\"eserviceId\":\"10000000109\",\"currentStatusHistoryId\":15000910007,\"hasTimestamp\":false,\"orderStatusName\":\"Заявление получено ведомством\",\"paymentRequired\":false,\"paymentStatusEvents\":[],\"payback\":false,\"readyToPush\":false,\"admLevelCode\":\"FEDERAL\",\"id\":3500308079,\"signCnt\":0,\"allFileSign\":false,\"noPaidPaymentCount\":-1,\"childrenSigned\":false,\"elk\":false,\"creationMode\":\"api\",\"steps\":[],\"orderStatusId\":2,\"withCustomResult\":false,\"statuses\":[{\"date\":\"2023-12-13T14:23:02.845+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"edit\",\"title\":\"Черновик заявления\",\"sendMessageAllowed\":false,\"statusId\":0,\"editAllowed\":false,\"id\":15000906353},{\"date\":\"2023-12-13T14:23:03.170+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Зарегистрировано на портале\",\"sendMessageAllowed\":false,\"statusId\":17,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000910006},{\"date\":\"2023-12-13T14:23:03.810+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление отправлено в ведомство\",\"sendMessageAllowed\":false,\"statusId\":21,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000900969},{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007}],\"orderDate\":\"2023-12-13T14:23:02.000+0300\",\"updated\":\"2023-12-13T14:23:11.434+0300\",\"hasNoPaidPayment\":false,\"servicePassportId\":\"600109\",\"checkQueue\":false,\"withDelivery\":false,\"gisdo\":false,\"userSelectedRegion\":\"00000000000\",\"sourceSystem\":\"774216\",\"eQueueEvents\":[],\"hasActiveInviteToEqueue\":false,\"multRegion\":true,\"serviceEpguId\":\"1\",\"extSystem\":false,\"useAsTemplate\":false,\"edsStatus\":\"EDS_NOT_SUPPORTED\",\"allowToDelete\":false,\"qrlink\":{\"hasAltMimeType\":false,\"fileSize\":0,\"canSentToMFC\":false,\"hasDigitalSignature\":false,\"canPrintMFC\":false},\"requestDate\":\"2023-12-13T14:23:03.175+0300\",\"hasPreviewPdf\":false,\"stateOrgCode\":\"pfr\",\"testUser\":false,\"personType\":\"PERSON\",\"textMessages\":[],\"serviceTargetId\":\"-10000000109\",\"orderPayments\":[],\"unreadMessageCnt\":0,\"orderResponseFiles\":[],\"hasResult\":false,\"serviceName\":\"Доставка пенсии и социальных выплат СФР\",\"deprecatedService\":false,\"hubForm\":false,\"userId\":1000571421,\"allowToEdit\":false,\"orderAttachmentFiles\":[{\"fileName\":\"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":4875,\"link\":\"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"REQUEST\"},{\"fileName\":\"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":604,\"link\":\"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"ATTACHMENT\"}],\"closed\":false,\"online\":false,\"readyToSign\":false,\"currentStatusHistory\":{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007},\"infoMessages\":[],\"location\":\"92000000000\",\"paymentCount\":0,\"draftHidden\":false,\"stateStructureId\":\"10000002796\"}"}`
const orderInfoOrderWantJSON = `{"orderType":"ORDER","hasNewStatus":true,"smevTx":"14329ae4-875b-40d3-9367-e2ef1b135189","stateOrgId":266,"hasEmpowerment2021":false,"smevMessageId":"WAIT_RESPONSE","hasChildren":false,"stateStructureName":"СФР","formVersion":"1","possibleServices":{},"orderAttributeEvents":[],"ownerId":1000571421,"eserviceId":"10000000109","currentStatusHistoryId":15000910007,"hasTimestamp":false,"orderStatusName":"Заявление получено ведомством","paymentRequired":false,"paymentStatusEvents":[],"payback":false,"readyToPush":false,"admLevelCode":"FEDERAL","id":3500308079,"signCnt":0,"allFileSign":false,"noPaidPaymentCount":-1,"childrenSigned":false,"elk":false,"creationMode":"api","steps":[],"orderStatusId":2,"withCustomResult":false,"statuses":[{"date":"2023-12-13T14:23:02.845+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"edit","title":"Черновик заявления","sendMessageAllowed":false,"statusId":0,"editAllowed":false,"id":15000906353},{"date":"2023-12-13T14:23:03.170+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Зарегистрировано на портале","sendMessageAllowed":false,"statusId":17,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000910006},{"date":"2023-12-13T14:23:03.810+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление отправлено в ведомство","sendMessageAllowed":false,"statusId":21,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000900969},{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007}],"orderDate":"2023-12-13T14:23:02.000+0300","updated":"2023-12-13T14:23:11.434+0300","hasNoPaidPayment":false,"servicePassportId":"600109","checkQueue":false,"withDelivery":false,"gisdo":false,"userSelectedRegion":"00000000000","sourceSystem":"774216","eQueueEvents":[],"hasActiveInviteToEqueue":false,"multRegion":true,"serviceEpguId":"1","extSystem":false,"useAsTemplate":false,"edsStatus":"EDS_NOT_SUPPORTED","allowToDelete":false,"qrlink":{"hasAltMimeType":false,"fileSize":0,"canSentToMFC":false,"hasDigitalSignature":false,"canPrintMFC":false},"requestDate":"2023-12-13T14:23:03.175+0300","hasPreviewPdf":false,"stateOrgCode":"pfr","testUser":false,"personType":"PERSON","textMessages":[],"serviceTargetId":"-10000000109","orderPayments":[],"unreadMessageCnt":0,"orderResponseFiles":[],"hasResult":false,"serviceName":"Доставка пенсии и социальных выплат СФР","deprecatedService":false,"hubForm":false,"userId":1000571421,"allowToEdit":false,"orderAttachmentFiles":[{"fileName":"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":4875,"link":"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw","mimeType":"application/xml","hasDigitalSignature":false,"type":"REQUEST"},{"fileName":"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":604,"link":"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA","mimeType":"application/xml","hasDigitalSignature":false,"type":"ATTACHMENT"}],"closed":false,"online":false,"readyToSign":false,"currentStatusHistory":{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007},"infoMessages":[],"location":"92000000000","paymentCount":0,"draftHidden":false,"stateStructureId":"10000002796"}`
