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

	suite.Run("200 success", func() {
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

	suite.Run("200 with unexpected json response", func() {
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

	suite.Run("200 with malformed json response", func() {
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
		suite.Equal("ошибка OrderCreate: ошибка чтения JSON: invalid character 'm' looking for beginning of value", err.Error())
		suite.Equal(0, orderId)
	})

	suite.Run("400 with no code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"Пользователь не дал согласие"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(testToken, testMeta)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrStatusBadRequest)
		suite.Equal(
			"ошибка OrderCreate: HTTP 400 Bad Request: неверные параметры: код ошибки не указан [code='', message='Пользователь не дал согласие']",
			err.Error(),
		)
		suite.Equal(0, orderId)
	})

	suite.Run("403 with access_denied_service", func() {
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

	suite.Run("200 success single chunk less than chunkSize", func() {
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

	suite.Run("200 success with single chunk equal to chunkSize", func() {
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

	suite.Run("200 success with multiple chunks", func() {
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

	suite.Run("200 success without orderId in response", func() {
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

	suite.Run("200 success with wrong orderId in response", func() {
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

	suite.Run("500 with unexpected plain text response", func() {
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

	suite.Run("200 success", func() {
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

	suite.Run("200 success without orderId in response", func() {
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

	suite.Run("409 with service_not_found code", func() {
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

	suite.Run("200 success", func() {
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

	suite.Run("200 success with null order field", func() {
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

	suite.Run("200 with malformed json response", func() {
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

	suite.Run("2oo with malformed order field", func() {
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

	suite.Run("204 order not found", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusOrderNotFound)
		suite.Equal("ошибка OrderInfo: HTTP 204 No Content: заявление не найдено", err.Error())
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

	suite.Run("401 unauthorized", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusUnauthorized)
		suite.Equal("ошибка OrderInfo: HTTP 401 Unauthorized: отказ в доступе", err.Error())
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
		suite.Equal("ошибка OrderInfo: HTTP 406 Not Acceptable: неожиданный HTTP-статус", err.Error())
		suite.Nil(orderInfo)
	})

	suite.Run("400 with unknown code", func() {
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
			"ошибка OrderInfo: HTTP 400 Bad Request: неверные параметры: неожиданный код ошибки [code='unknown_code', message='Unknown Code']",
			err.Error(),
		)
		suite.Nil(orderInfo)
	})
}

func (suite *suiteTestClient) TestOrderCancel() {

	suite.Run("200 success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/gusmev/order/123456/cancel", r.URL.Path)
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(testToken, 123456)
		suite.NoError(err)
	})

	suite.Run("404 not found", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`Not Found`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrStatusURLNotFound)
		suite.Equal(
			"ошибка OrderCancel: HTTP 404 Not Found: не найден URL запроса",
			err.Error(),
		)
	})

	suite.Run("429 too many requests", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrStatusTooManyRequests)
		suite.Equal(
			"ошибка OrderCancel: HTTP 429 Too Many Requests: слишком много запросов",
			err.Error(),
		)
	})

	suite.Run("502 with plain text response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("bad gateway"))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrStatusBadGateway)
		suite.Equal(
			"ошибка OrderCancel: HTTP 502 Bad Gateway: некорректный шлюз",
			err.Error(),
		)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		err := client.OrderCancel(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrRequest)
	})

	suite.Run("409 with cancel_not_allowed code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"cancel_not_allowed", "message":"Заявление не может быть отменено"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrStatusUnableToHandleRequest)
		suite.ErrorIs(err, ErrCodeCancelNotAllowed)
		suite.Equal(
			"ошибка OrderCancel: HTTP 409 Conflict: невозможно обработать запрос: отмена заявления в текущем статусе невозможна [code='cancel_not_allowed', message='Заявление не может быть отменено']",
			err.Error(),
		)
	})

}

func (suite *suiteTestClient) Test_attachmentURI() {
	suite.Run("normal link", func() {
		uri, err := attachmentURI("terrabyte://00/12345678/req_some-guid-1234.xml/2")
		suite.NoError(err)
		suite.Equal("/12345678/2/download?mnemonic=req_some-guid-1234.xml", uri)
	})

	suite.Run("some prefix", func() {
		uri, err := attachmentURI("terrabyte://00/some/more/12345678/req_some-guid-1234.xml/2")
		suite.NoError(err)
		suite.Equal("/12345678/2/download?mnemonic=req_some-guid-1234.xml", uri)
	})

	suite.Run("no mnemonic", func() {
		uri, err := attachmentURI("terrabyte://00/12345678/2")
		suite.ErrorIs(err, ErrInvalidFileLink)
		suite.Equal("", uri)
	})

}

var (
	testToken = "test-token"
	testMeta  = OrderMeta{Region: "test-region", ServiceCode: "test-service", TargetCode: "test-target"}
)

const orderInfoResponseJSON = `{"code":"OK","message":"test","messageId":"test-GUID","order":"{\"orderType\":\"ORDER\",\"hasNewStatus\":true,\"smevTx\":\"14329ae4-875b-40d3-9367-e2ef1b135189\",\"stateOrgId\":266,\"hasEmpowerment2021\":false,\"smevMessageId\":\"WAIT_RESPONSE\",\"hasChildren\":false,\"stateStructureName\":\"СФР\",\"formVersion\":\"1\",\"possibleServices\":{},\"orderAttributeEvents\":[],\"ownerId\":1000571421,\"eserviceId\":\"10000000109\",\"currentStatusHistoryId\":15000910007,\"hasTimestamp\":false,\"orderStatusName\":\"Заявление получено ведомством\",\"paymentRequired\":false,\"paymentStatusEvents\":[],\"payback\":false,\"readyToPush\":false,\"admLevelCode\":\"FEDERAL\",\"id\":3500308079,\"signCnt\":0,\"allFileSign\":false,\"noPaidPaymentCount\":-1,\"childrenSigned\":false,\"elk\":false,\"creationMode\":\"api\",\"steps\":[],\"orderStatusId\":2,\"withCustomResult\":false,\"statuses\":[{\"date\":\"2023-12-13T14:23:02.845+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"edit\",\"title\":\"Черновик заявления\",\"sendMessageAllowed\":false,\"statusId\":0,\"editAllowed\":false,\"id\":15000906353},{\"date\":\"2023-12-13T14:23:03.170+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Зарегистрировано на портале\",\"sendMessageAllowed\":false,\"statusId\":17,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000910006},{\"date\":\"2023-12-13T14:23:03.810+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление отправлено в ведомство\",\"sendMessageAllowed\":false,\"statusId\":21,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000900969},{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007}],\"orderDate\":\"2023-12-13T14:23:02.000+0300\",\"updated\":\"2023-12-13T14:23:11.434+0300\",\"hasNoPaidPayment\":false,\"servicePassportId\":\"600109\",\"checkQueue\":false,\"withDelivery\":false,\"gisdo\":false,\"userSelectedRegion\":\"00000000000\",\"sourceSystem\":\"774216\",\"eQueueEvents\":[],\"hasActiveInviteToEqueue\":false,\"multRegion\":true,\"serviceEpguId\":\"1\",\"extSystem\":false,\"useAsTemplate\":false,\"edsStatus\":\"EDS_NOT_SUPPORTED\",\"allowToDelete\":false,\"qrlink\":{\"hasAltMimeType\":false,\"fileSize\":0,\"canSentToMFC\":false,\"hasDigitalSignature\":false,\"canPrintMFC\":false},\"requestDate\":\"2023-12-13T14:23:03.175+0300\",\"hasPreviewPdf\":false,\"stateOrgCode\":\"pfr\",\"testUser\":false,\"personType\":\"PERSON\",\"textMessages\":[],\"serviceTargetId\":\"-10000000109\",\"orderPayments\":[],\"unreadMessageCnt\":0,\"orderResponseFiles\":[],\"hasResult\":false,\"serviceName\":\"Доставка пенсии и социальных выплат СФР\",\"deprecatedService\":false,\"hubForm\":false,\"userId\":1000571421,\"allowToEdit\":false,\"orderAttachmentFiles\":[{\"fileName\":\"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":4875,\"link\":\"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"REQUEST\"},{\"fileName\":\"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":604,\"link\":\"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"ATTACHMENT\"}],\"closed\":false,\"online\":false,\"readyToSign\":false,\"currentStatusHistory\":{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007},\"infoMessages\":[],\"location\":\"92000000000\",\"paymentCount\":0,\"draftHidden\":false,\"stateStructureId\":\"10000002796\"}"}`
const orderInfoOrderWantJSON = `{"orderType":"ORDER","hasNewStatus":true,"smevTx":"14329ae4-875b-40d3-9367-e2ef1b135189","stateOrgId":266,"hasEmpowerment2021":false,"smevMessageId":"WAIT_RESPONSE","hasChildren":false,"stateStructureName":"СФР","formVersion":"1","possibleServices":{},"orderAttributeEvents":[],"ownerId":1000571421,"eserviceId":"10000000109","currentStatusHistoryId":15000910007,"hasTimestamp":false,"orderStatusName":"Заявление получено ведомством","paymentRequired":false,"paymentStatusEvents":[],"payback":false,"readyToPush":false,"admLevelCode":"FEDERAL","id":3500308079,"signCnt":0,"allFileSign":false,"noPaidPaymentCount":-1,"childrenSigned":false,"elk":false,"creationMode":"api","steps":[],"orderStatusId":2,"withCustomResult":false,"statuses":[{"date":"2023-12-13T14:23:02.845+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"edit","title":"Черновик заявления","sendMessageAllowed":false,"statusId":0,"editAllowed":false,"id":15000906353},{"date":"2023-12-13T14:23:03.170+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Зарегистрировано на портале","sendMessageAllowed":false,"statusId":17,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000910006},{"date":"2023-12-13T14:23:03.810+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление отправлено в ведомство","sendMessageAllowed":false,"statusId":21,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000900969},{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007}],"orderDate":"2023-12-13T14:23:02.000+0300","updated":"2023-12-13T14:23:11.434+0300","hasNoPaidPayment":false,"servicePassportId":"600109","checkQueue":false,"withDelivery":false,"gisdo":false,"userSelectedRegion":"00000000000","sourceSystem":"774216","eQueueEvents":[],"hasActiveInviteToEqueue":false,"multRegion":true,"serviceEpguId":"1","extSystem":false,"useAsTemplate":false,"edsStatus":"EDS_NOT_SUPPORTED","allowToDelete":false,"qrlink":{"hasAltMimeType":false,"fileSize":0,"canSentToMFC":false,"hasDigitalSignature":false,"canPrintMFC":false},"requestDate":"2023-12-13T14:23:03.175+0300","hasPreviewPdf":false,"stateOrgCode":"pfr","testUser":false,"personType":"PERSON","textMessages":[],"serviceTargetId":"-10000000109","orderPayments":[],"unreadMessageCnt":0,"orderResponseFiles":[],"hasResult":false,"serviceName":"Доставка пенсии и социальных выплат СФР","deprecatedService":false,"hubForm":false,"userId":1000571421,"allowToEdit":false,"orderAttachmentFiles":[{"fileName":"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":4875,"link":"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw","mimeType":"application/xml","hasDigitalSignature":false,"type":"REQUEST"},{"fileName":"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":604,"link":"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA","mimeType":"application/xml","hasDigitalSignature":false,"type":"ATTACHMENT"}],"closed":false,"online":false,"readyToSign":false,"currentStatusHistory":{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007},"infoMessages":[],"location":"92000000000","paymentCount":0,"draftHidden":false,"stateStructureId":"10000002796"}`
