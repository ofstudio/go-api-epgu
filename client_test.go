package apipgu

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
		suite.NoError(err)
		suite.Equal(123456, orderId)
	})

	suite.Run("200 success with escaped meta", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			suite.JSONEq(`{"region":"test \"region\"","serviceCode":"test\\service","targetCode":"test\ntarget"}`, string(body))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":123456}`))
		}))
		defer server.Close()

		meta := OrderMeta{Region: `test "region"`, ServiceCode: `test\service`, TargetCode: "test\ntarget"}
		client := NewClient(server.URL)
		orderId, err := client.OrderCreate(context.Background(), testToken, meta)
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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
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
		orderId, err := client.OrderCreate(context.Background(), testToken, testMeta)
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
			_, found := r.Form["chunk"]
			suite.False(found)
			_, found = r.Form["chunks"]
			suite.False(found)
			suite.Equal("123456", r.FormValue("orderId"))
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
		suite.NoError(client.OrderPushChunked(context.Background(), testToken, 123456, testArchive))
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
			suite.Equal(fmt.Sprintf("%d", reqCount-1), r.FormValue("chunk"))
			suite.Equal("4", r.FormValue("chunks"))
			suite.Equal("123456", r.FormValue("orderId"))
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
		suite.NoError(client.OrderPushChunked(context.Background(), testToken, 123456, testArchive))
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
		suite.NoError(err)
	})

	suite.Run("archive is nil", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: nil}
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrNilArchive)
	})

	suite.Run("archive is zero length", func() {
		client := NewClient("").WithChunkSize(100)
		testArchive := &Archive{Name: "test-archive", Data: []byte{}}
		err := client.OrderPushChunked(context.Background(), testToken, 123456, testArchive)
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrRequest)
		suite.Equal(0, orderId)
	})

	suite.Run("archive is nil", func() {
		client := NewClient("")
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, nil)
		suite.Error(err)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrNilArchive)
		suite.Equal(0, orderId)
	})

	suite.Run("archive is zero length", func() {
		client := NewClient("")
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, &Archive{})
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
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, testArchive)
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
			_, _ = w.Write([]byte(orderInfoSuccessResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
		suite.NoError(err)
		suite.NotNil(orderInfo)
		suite.Equal("OK", orderInfo.Code)
		suite.Equal("test", orderInfo.Message)
		suite.Equal("test-GUID", orderInfo.MessageId)
		suite.NotNil(orderInfo.Order)
		orderJSON, err := json.Marshal(orderInfo.Order)
		suite.NoError(err)
		suite.JSONEq(orderInfoSuccessWant, string(orderJSON))
	})

	suite.Run("200 success with null order field", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":"OK","message":"test","messageId":"test-GUID","order":null}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
		suite.Error(err)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrStatusOrderNotFound)
		suite.Equal("ошибка OrderInfo: HTTP 204 No Content: заявление не найдено", err.Error())
		suite.Nil(orderInfo)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
		orderInfo, err := client.OrderInfo(context.Background(), testToken, 123456)
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
			suite.Empty(r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.OrderCancel(context.Background(), testToken, 123456)
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
		err := client.OrderCancel(context.Background(), testToken, 123456)
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
		err := client.OrderCancel(context.Background(), testToken, 123456)
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
		err := client.OrderCancel(context.Background(), testToken, 123456)
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
		err := client.OrderCancel(context.Background(), testToken, 123456)
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
		err := client.OrderCancel(context.Background(), testToken, 123456)
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

func (suite *suiteTestClient) TestGetOrdersStatus() {
	suite.Run("200 success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodGet, r.Method)
			suite.Equal("/api/gusmev/order/getOrdersStatus?pageNum=0&pageSize=5&orderIds=764607248,764603930,2354270898", r.RequestURI)
			suite.Equal("application/json", r.Header.Get("Content-Type"))
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(ordersStatusSuccessResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ordersStatus, err := client.GetOrdersStatus(context.Background(), testToken, 0, 5, []int{764607248, 764603930, 2354270898})
		suite.NoError(err)
		suite.Require().NotNil(ordersStatus)
		suite.Equal(3, ordersStatus.Count)
		suite.Equal(5, ordersStatus.TotalCount)
		suite.Len(ordersStatus.Content, 3)
		suite.Equal(764607248, ordersStatus.Content[0].OrderId)
		suite.Equal("FOUND", ordersStatus.Content[0].OrderSearchStatus)
		suite.Equal(24, ordersStatus.Content[0].Status.StatusId)
		suite.Equal("Ошибка отправки заявления в ведомство", ordersStatus.Content[0].Status.StatusName)
		suite.Equal(time.Date(2022, 12, 21, 20, 49, 37, 672000000, time.UTC), ordersStatus.Content[0].Status.Updated.Time)
		suite.Equal(2354270898, ordersStatus.Content[2].OrderId)
		suite.Equal("NOT_FOUND", ordersStatus.Content[2].OrderSearchStatus)
		suite.True(ordersStatus.Content[2].Status.Updated.Time.IsZero())
	})

	suite.Run("400 with bad_request code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":"bad_request", "message":"Некорректные параметры"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ordersStatus, err := client.GetOrdersStatus(context.Background(), testToken, 0, 5, []int{764607248})
		suite.Error(err)
		suite.ErrorIs(err, ErrGetOrdersStatus)
		suite.ErrorIs(err, ErrStatusBadRequest)
		suite.ErrorIs(err, ErrCodeBadRequest)
		suite.Nil(ordersStatus)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		ordersStatus, err := client.GetOrdersStatus(context.Background(), testToken, 0, 5, []int{764607248})
		suite.Error(err)
		suite.ErrorIs(err, ErrGetOrdersStatus)
		suite.ErrorIs(err, ErrRequest)
		suite.Nil(ordersStatus)
	})
}

func (suite *suiteTestClient) TestGetUpdatedAfter() {
	suite.Run("200 success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodGet, r.Method)
			suite.Equal("/api/gusmev/order/getUpdatedAfter?pageNum=0&pageSize=5&updatedAfter=2022-12-10T12:31:42.000", r.RequestURI)
			suite.Empty(r.Header.Get("Content-Type"))
			suite.Equal("Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(ordersStatusSuccessResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		updatedAfter := time.Date(2022, 12, 10, 12, 31, 42, 0, time.UTC)
		ordersStatus, err := client.GetUpdatedAfter(context.Background(), testToken, 0, 5, updatedAfter)
		suite.NoError(err)
		suite.Require().NotNil(ordersStatus)
		suite.Equal(3, ordersStatus.Count)
		suite.Equal(5, ordersStatus.TotalCount)
		suite.Len(ordersStatus.Content, 3)
	})

	suite.Run("500 with malformed json response", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`malformed json{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ordersStatus, err := client.GetUpdatedAfter(context.Background(), testToken, 0, 5, time.Now())
		suite.Error(err)
		suite.ErrorIs(err, ErrGetUpdatedAfter)
		suite.ErrorIs(err, ErrStatusInternalError)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Nil(ordersStatus)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		ordersStatus, err := client.GetUpdatedAfter(context.Background(), testToken, 0, 5, time.Now())
		suite.Error(err)
		suite.ErrorIs(err, ErrGetUpdatedAfter)
		suite.ErrorIs(err, ErrRequest)
		suite.Nil(ordersStatus)
	})
}

func (suite *suiteTestClient) TestAttachmentDownload() {

	suite.Run("200 success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodGet, r.Method)
			suite.Equal("/api/gusmev/files/download/15000910007/3", r.URL.Path)
			suite.Equal("result.xml", r.URL.Query().Get("mnemonic"))
			suite.Equal("10000000109", r.URL.Query().Get("eserviceCode"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test data"))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		body := &bytes.Buffer{}
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "terrabyte://00/3500308079/result.xml/3", "10000000109", body)
		suite.NoError(err)
		suite.Equal("test data", body.String())
	})

	suite.Run("404 not found", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrStatusURLNotFound)
		suite.Equal("ошибка AttachmentDownload: HTTP 404 Not Found: не найден URL запроса", err.Error())
	})

	suite.Run("503 service unavailable", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrStatusServiceUnavailable)
		suite.Equal("ошибка AttachmentDownload: HTTP 503 Service Unavailable: сервис недоступен", err.Error())
	})

	suite.Run("403 with access_denied_user code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code":"access_denied_user", "message":"Доступ запрещен"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrStatusForbidden)
		suite.ErrorIs(err, ErrCodeAccessDeniedUser)
		suite.Equal("ошибка AttachmentDownload: HTTP 403 Forbidden: доступ запрещен: доступ запрещен для данного типа пользователя [code='access_denied_user', message='Доступ запрещен']", err.Error())
	})

	suite.Run("invalid file link", func() {
		client := NewClient("")
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "invalid link", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrInvalidFileLink)
	})

	suite.Run("invalid object id", func() {
		client := NewClient("")
		err := client.AttachmentDownload(context.Background(), testToken, 0, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrInvalidFileLink)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		err := client.AttachmentDownload(context.Background(), testToken, 15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.Error(err)
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrRequest)
	})

}
func (suite *suiteTestClient) Test_attachmentEndpoint() {
	suite.Run("normal link", func() {
		uri, err := attachmentEndpoint(15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109")
		suite.NoError(err)
		suite.Equal("/api/gusmev/files/download/15000910007/2?eserviceCode=10000000109&mnemonic=req_some-guid-1234.xml", uri)
	})

	suite.Run("some prefix", func() {
		uri, err := attachmentEndpoint(15000910007, "terrabyte://00/some/more/3500308079/req_some-guid-1234.xml/2", "10000000109")
		suite.NoError(err)
		suite.Equal("/api/gusmev/files/download/15000910007/2?eserviceCode=10000000109&mnemonic=req_some-guid-1234.xml", uri)
	})

	suite.Run("escape path and query", func() {
		uri, err := attachmentEndpoint(15000910007, "terrabyte://00/3500308079/req some.xml/тип 3", "10000000109/test")
		suite.NoError(err)
		suite.Equal("/api/gusmev/files/download/15000910007/%D1%82%D0%B8%D0%BF%203?eserviceCode=10000000109%2Ftest&mnemonic=req+some.xml", uri)
	})

	suite.Run("no mnemonic", func() {
		uri, err := attachmentEndpoint(15000910007, "terrabyte://00/12345678//2", "10000000109")
		suite.ErrorIs(err, ErrInvalidFileLink)
		suite.Equal("", uri)
	})

}

func (suite *suiteTestClient) TestContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	archive := &Archive{Name: "test-archive", Data: []byte("test")}

	suite.Run("OrderCreate", func() {
		client := NewClient("http://127.0.0.1")
		orderId, err := client.OrderCreate(ctx, testToken, testMeta)
		suite.ErrorIs(err, ErrOrderCreate)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Equal(0, orderId)
	})

	suite.Run("OrderPushChunked", func() {
		client := NewClient("http://127.0.0.1")
		err := client.OrderPushChunked(ctx, testToken, 123456, archive)
		suite.ErrorIs(err, ErrPushChunked)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
	})

	suite.Run("OrderPush", func() {
		client := NewClient("http://127.0.0.1")
		orderId, err := client.OrderPush(ctx, testToken, testMeta, archive)
		suite.ErrorIs(err, ErrPush)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Equal(0, orderId)
	})

	suite.Run("OrderInfo", func() {
		client := NewClient("http://127.0.0.1")
		orderInfo, err := client.OrderInfo(ctx, testToken, 123456)
		suite.ErrorIs(err, ErrOrderInfo)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Nil(orderInfo)
	})

	suite.Run("OrderCancel", func() {
		client := NewClient("http://127.0.0.1")
		err := client.OrderCancel(ctx, testToken, 123456)
		suite.ErrorIs(err, ErrOrderCancel)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
	})

	suite.Run("GetOrdersStatus", func() {
		client := NewClient("http://127.0.0.1")
		ordersStatus, err := client.GetOrdersStatus(ctx, testToken, 0, 5, []int{123456})
		suite.ErrorIs(err, ErrGetOrdersStatus)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Nil(ordersStatus)
	})

	suite.Run("GetUpdatedAfter", func() {
		client := NewClient("http://127.0.0.1")
		ordersStatus, err := client.GetUpdatedAfter(ctx, testToken, 0, 5, time.Now())
		suite.ErrorIs(err, ErrGetUpdatedAfter)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Nil(ordersStatus)
	})

	suite.Run("AttachmentDownload", func() {
		client := NewClient("http://127.0.0.1")
		err := client.AttachmentDownload(ctx, testToken, 15000910007, "terrabyte://00/3500308079/req_some-guid-1234.xml/2", "10000000109", &bytes.Buffer{})
		suite.ErrorIs(err, ErrAttachmentDownload)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
	})

	suite.Run("Dict", func() {
		client := NewClient("http://127.0.0.1")
		items, n, err := client.Dict(ctx, "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.ErrorIs(err, ErrDict)
		suite.ErrorIs(err, ErrRequest)
		suite.ErrorIs(err, context.Canceled)
		suite.Equal(0, n)
		suite.Nil(items)
	})
}

func (suite *suiteTestClient) TestDict() {

	suite.Run("200 success with simple dict", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)
			body, err := io.ReadAll(r.Body)
			suite.NoError(err)
			suite.JSONEq(`{"treeFiltering":"SUBTREE"}`, string(body))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dictSuccessSimpleResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.NoError(err)
		suite.Equal(5004, n)
		suite.Len(items, 2)
		itemsJSON, err := json.Marshal(items)
		suite.NoError(err)
		suite.JSONEq(dictSuccessSimpleWant, string(itemsJSON))
	})

	suite.Run("200 success with escaped code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal("/api/nsi/v1/dictionary/TEST%2FDICT", r.RequestURI)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dictSuccessEmptyResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST/DICT", DictFilterSubTree, "", 0, 0)
		suite.NoError(err)
		suite.Equal(5004, n)
		suite.Len(items, 0)
	})

	suite.Run("200 success with pagination and complex dict", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)
			body, err := io.ReadAll(r.Body)
			suite.NoError(err)
			suite.JSONEq(`{"pageNum":10, "pageSize":20, "parentRefItemValue":"test_parent", "treeFiltering":"ONELEVEL"}`, string(body))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dictSuccessComplexResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterOneLevel, "test_parent", 10, 20)
		suite.NoError(err)
		suite.Equal(1000, n)
		suite.Len(items, 2)
		itemsJSON, err := json.Marshal(items)
		suite.NoError(err)
		suite.JSONEq(dictSuccessComplexWant, string(itemsJSON))
	})

	suite.Run("200 with empty result", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dictSuccessEmptyResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", "", DictFilterSubTree, 0, 0)
		suite.NoError(err)
		suite.Equal(5004, n)
		suite.Len(items, 0)
	})

	suite.Run("200 with error code", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dictErrorResponse))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.Error(err)
		suite.ErrorIs(err, ErrDict)
		suite.ErrorIs(err, ErrDictResponse)
		suite.Equal(
			"ошибка Dict: ошибка получения справочных данных [code='7', message='Entity not found']",
			err.Error(),
		)
		suite.Equal(0, n)
		suite.Len(items, 0)
	})

	suite.Run("200 with malformed json", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`malformed json{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.Error(err)
		suite.ErrorIs(err, ErrDict)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(
			"ошибка Dict: ошибка чтения JSON: invalid character 'm' looking for beginning of value",
			err.Error(),
		)
		suite.Equal(0, n)
		suite.Len(items, 0)
	})

	suite.Run("504 with no content", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/api/nsi/v1/dictionary/TEST_DICT", r.URL.Path)

			w.WriteHeader(http.StatusGatewayTimeout)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.Error(err)
		suite.ErrorIs(err, ErrDict)
		suite.ErrorIs(err, ErrStatusGatewayTimeout)
		suite.Equal("ошибка Dict: HTTP 504 Gateway Timeout: шлюз не отвечает", err.Error())
		suite.Equal(0, n)
		suite.Len(items, 0)
	})

	suite.Run("request error", func() {
		client := NewClient("")
		items, n, err := client.Dict(context.Background(), "TEST_DICT", DictFilterSubTree, "", 0, 0)
		suite.Error(err)
		suite.ErrorIs(err, ErrDict)
		suite.ErrorIs(err, ErrRequest)
		suite.Equal(0, n)
		suite.Len(items, 0)
	})

}

var (
	testToken = "test-token"
	testMeta  = OrderMeta{Region: "test-region", ServiceCode: "test-service", TargetCode: "test-target"}
)

const (
	orderInfoSuccessResponse = `{"code":"OK","message":"test","messageId":"test-GUID","order":"{\"orderType\":\"ORDER\",\"hasNewStatus\":true,\"smevTx\":\"14329ae4-875b-40d3-9367-e2ef1b135189\",\"stateOrgId\":266,\"hasEmpowerment2021\":false,\"smevMessageId\":\"WAIT_RESPONSE\",\"hasChildren\":false,\"stateStructureName\":\"СФР\",\"formVersion\":\"1\",\"possibleServices\":{},\"orderAttributeEvents\":[],\"ownerId\":1000571421,\"eserviceId\":\"10000000109\",\"currentStatusHistoryId\":15000910007,\"hasTimestamp\":false,\"orderStatusName\":\"Заявление получено ведомством\",\"paymentRequired\":false,\"paymentStatusEvents\":[],\"payback\":false,\"readyToPush\":false,\"admLevelCode\":\"FEDERAL\",\"id\":3500308079,\"signCnt\":0,\"allFileSign\":false,\"noPaidPaymentCount\":-1,\"childrenSigned\":false,\"elk\":false,\"creationMode\":\"api\",\"steps\":[],\"orderStatusId\":2,\"withCustomResult\":false,\"statuses\":[{\"date\":\"2023-12-13T14:23:02.845+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"edit\",\"title\":\"Черновик заявления\",\"sendMessageAllowed\":false,\"statusId\":0,\"editAllowed\":false,\"id\":15000906353},{\"date\":\"2023-12-13T14:23:03.170+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Зарегистрировано на портале\",\"sendMessageAllowed\":false,\"statusId\":17,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000910006},{\"date\":\"2023-12-13T14:23:03.810+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление отправлено в ведомство\",\"sendMessageAllowed\":false,\"statusId\":21,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"id\":15000900969},{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007}],\"orderDate\":\"2023-12-13T14:23:02.000+0300\",\"updated\":\"2023-12-13T14:23:11.434+0300\",\"hasNoPaidPayment\":false,\"servicePassportId\":\"600109\",\"checkQueue\":false,\"withDelivery\":false,\"gisdo\":false,\"userSelectedRegion\":\"00000000000\",\"sourceSystem\":\"774216\",\"eQueueEvents\":[],\"hasActiveInviteToEqueue\":false,\"multRegion\":true,\"serviceEpguId\":\"1\",\"extSystem\":false,\"useAsTemplate\":false,\"edsStatus\":\"EDS_NOT_SUPPORTED\",\"allowToDelete\":false,\"qrlink\":{\"hasAltMimeType\":false,\"fileSize\":0,\"canSentToMFC\":false,\"hasDigitalSignature\":false,\"canPrintMFC\":false},\"requestDate\":\"2023-12-13T14:23:03.175+0300\",\"hasPreviewPdf\":false,\"stateOrgCode\":\"pfr\",\"testUser\":false,\"personType\":\"PERSON\",\"textMessages\":[],\"serviceTargetId\":\"-10000000109\",\"orderPayments\":[],\"unreadMessageCnt\":0,\"orderResponseFiles\":[],\"hasResult\":false,\"serviceName\":\"Доставка пенсии и социальных выплат СФР\",\"deprecatedService\":false,\"hubForm\":false,\"userId\":1000571421,\"allowToEdit\":false,\"orderAttachmentFiles\":[{\"fileName\":\"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":4875,\"link\":\"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"REQUEST\"},{\"fileName\":\"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml\",\"fileSize\":604,\"link\":\"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2\",\"id\":\"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"ATTACHMENT\"}],\"closed\":false,\"online\":false,\"readyToSign\":false,\"currentStatusHistory\":{\"date\":\"2023-12-13T14:23:11.429+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":3500308079,\"hasResult\":\"N\",\"statusColorCode\":\"in_progress\",\"title\":\"Заявление получено ведомством\",\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Фонд пенсионного и социального страхования Российской Федерации\",\"comment\":\"Сообщение доставлено\",\"id\":15000910007},\"infoMessages\":[],\"location\":\"92000000000\",\"paymentCount\":0,\"draftHidden\":false,\"stateStructureId\":\"10000002796\"}"}`
	orderInfoSuccessWant     = `{"orderType":"ORDER","hasNewStatus":true,"smevTx":"14329ae4-875b-40d3-9367-e2ef1b135189","stateOrgId":266,"hasEmpowerment2021":false,"smevMessageId":"WAIT_RESPONSE","hasChildren":false,"stateStructureName":"СФР","formVersion":"1","possibleServices":{},"orderAttributeEvents":[],"ownerId":1000571421,"eserviceId":"10000000109","currentStatusHistoryId":15000910007,"hasTimestamp":false,"orderStatusName":"Заявление получено ведомством","paymentRequired":false,"paymentStatusEvents":[],"payback":false,"readyToPush":false,"admLevelCode":"FEDERAL","id":3500308079,"signCnt":0,"allFileSign":false,"noPaidPaymentCount":-1,"childrenSigned":false,"elk":false,"creationMode":"api","steps":[],"orderStatusId":2,"withCustomResult":false,"statuses":[{"date":"2023-12-13T14:23:02.845+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"edit","title":"Черновик заявления","sendMessageAllowed":false,"statusId":0,"editAllowed":false,"id":15000906353},{"date":"2023-12-13T14:23:03.170+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Зарегистрировано на портале","sendMessageAllowed":false,"statusId":17,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000910006},{"date":"2023-12-13T14:23:03.810+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление отправлено в ведомство","sendMessageAllowed":false,"statusId":21,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","id":15000900969},{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007}],"orderDate":"2023-12-13T14:23:02.000+0300","updated":"2023-12-13T14:23:11.434+0300","hasNoPaidPayment":false,"servicePassportId":"600109","checkQueue":false,"withDelivery":false,"gisdo":false,"userSelectedRegion":"00000000000","sourceSystem":"774216","eQueueEvents":[],"hasActiveInviteToEqueue":false,"multRegion":true,"serviceEpguId":"1","extSystem":false,"useAsTemplate":false,"edsStatus":"EDS_NOT_SUPPORTED","allowToDelete":false,"qrlink":{"hasAltMimeType":false,"fileSize":0,"canSentToMFC":false,"hasDigitalSignature":false,"canPrintMFC":false},"requestDate":"2023-12-13T14:23:03.175+0300","hasPreviewPdf":false,"stateOrgCode":"pfr","testUser":false,"personType":"PERSON","textMessages":[],"serviceTargetId":"-10000000109","orderPayments":[],"unreadMessageCnt":0,"orderResponseFiles":[],"hasResult":false,"serviceName":"Доставка пенсии и социальных выплат СФР","deprecatedService":false,"hubForm":false,"userId":1000571421,"allowToEdit":false,"orderAttachmentFiles":[{"fileName":"req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":4875,"link":"terrabyte://00/3500308079/req_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/cmVxXzMwZWY5MzYyLTc2ZjAtNGE3Yi05YTBmLWYzYmE0M2MzNTRkNi54bWw","mimeType":"application/xml","hasDigitalSignature":false,"type":"REQUEST"},{"fileName":"trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml","fileSize":604,"link":"terrabyte://00/3500308079/trans_30ef9362-76f0-4a7b-9a0f-f3ba43c354d6.xml/2","id":"3500308079/files/dHJhbnNfMzBlZjkzNjItNzZmMC00YTdiLTlhMGYtZjNiYTQzYzM1NGQ2LnhtbA","mimeType":"application/xml","hasDigitalSignature":false,"type":"ATTACHMENT"}],"closed":false,"online":false,"readyToSign":false,"currentStatusHistory":{"date":"2023-12-13T14:23:11.429+0300","cancelAllowed":false,"unreadEvent":true,"deliveryCancelAllowed":false,"finalStatus":false,"orderId":3500308079,"hasResult":"N","statusColorCode":"in_progress","title":"Заявление получено ведомством","sendMessageAllowed":false,"statusId":2,"editAllowed":false,"sender":"Фонд пенсионного и социального страхования Российской Федерации","comment":"Сообщение доставлено","id":15000910007},"infoMessages":[],"location":"92000000000","paymentCount":0,"draftHidden":false,"stateStructureId":"10000002796"}`
)

const (
	ordersStatusSuccessResponse = `{
		"count": 3,
		"totalCount": 5,
		"content": [
			{
				"orderId": 764607248,
				"orderSearchStatus": "FOUND",
				"status": {
					"statusId": 24,
					"statusName": "Ошибка отправки заявления в ведомство",
					"updated": "2022-12-21T20:49:37.672"
				}
			},
			{
				"orderId": 764603930,
				"orderSearchStatus": "FOUND",
				"status": {
					"statusId": 15,
					"statusName": "Заявление требует исправления",
					"updated": "2022-12-21T19:51:16.852"
				}
			},
			{
				"orderId": 2354270898,
				"orderSearchStatus": "NOT_FOUND",
				"status": null
			}
		]
	}`

	dictSuccessSimpleResponse  = `{"error":{"code":0,"message":"operation completed"},"fieldErrors":[],"total":5004,"items":[{"value":"0550041","title":"1.Клиентская служба (на правах отдела) в Белозерском районе","isLeaf":true,"children":[],"attributes":[],"attributeValues":{}},{"value":"0550091","title":"1. Клиентская служба (на правах  отдела) в Лебяжьевском районе","isLeaf":true,"children":[],"attributes":[],"attributeValues":{}}]}`
	dictSuccessSimpleWant      = `[{"value":"0550041","title":"1.Клиентская служба (на правах отдела) в Белозерском районе","isLeaf":true,"children":[],"attributes":[],"attributeValues":{}},{"value":"0550091","title":"1. Клиентская служба (на правах  отдела) в Лебяжьевском районе","isLeaf":true,"children":[],"attributes":[],"attributeValues":{}}]`
	dictSuccessComplexResponse = `{"error":{"code":0,"message":"operation completed"},"fieldErrors":[],"total":1000,"items":[ {"value": "049514608", "title": "049514608 - АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан", "isLeaf": true, "children": [], "attributes": [ { "name": "ID", "type": "STRING", "value": { "asString": "049514608", "typeOfValue": "STRING", "value": "049514608" }, "valueAsOfType": "049514608" }, { "name": "NAME", "type": "STRING", "value": { "asString": "АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан", "typeOfValue": "STRING", "value": "АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан" }, "valueAsOfType": "АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан" }, { "name": "BIC", "type": "STRING", "value": { "asString": "049514608", "typeOfValue": "STRING", "value": "049514608" }, "valueAsOfType": "049514608" }, { "name": "CORR_ACCOUNT", "type": "STRING", "value": { "asString": "30101810500000000608", "typeOfValue": "STRING", "value": "30101810500000000608" }, "valueAsOfType": "30101810500000000608" } ], "attributeValues": { "ID": "049514608", "CORR_ACCOUNT": "30101810500000000608", "BIC": "049514608", "NAME": "АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан" } }, { "value": "041012765", "title": "041012765 - \"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск", "isLeaf": true, "children": [], "attributes": [ { "name": "ID", "type": "STRING", "value": { "asString": "041012765", "typeOfValue": "STRING", "value": "041012765" }, "valueAsOfType": "041012765" }, { "name": "NAME", "type": "STRING", "value": { "asString": "\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск", "typeOfValue": "STRING", "value": "\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск" }, "valueAsOfType": "\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск" }, { "name": "BIC", "type": "STRING", "value": { "asString": "041012765", "typeOfValue": "STRING", "value": "041012765" }, "valueAsOfType": "041012765" }, { "name": "CORR_ACCOUNT", "type": "STRING", "value": { "asString": "30101810300000000765", "typeOfValue": "STRING", "value": "30101810300000000765" }, "valueAsOfType": "30101810300000000765" } ], "attributeValues": { "ID": "041012765", "CORR_ACCOUNT": "30101810300000000765", "BIC": "041012765", "NAME": "\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск" } }]}`
	dictSuccessComplexWant     = `[{"value":"049514608","title":"049514608 - АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан","isLeaf":true,"children":[],"attributes":[{"name":"ID","type":"STRING","value":{"asString":"049514608","typeOfValue":"STRING","value":"049514608"},"valueAsOfType":"049514608"},{"name":"NAME","type":"STRING","value":{"asString":"АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан","typeOfValue":"STRING","value":"АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан"},"valueAsOfType":"АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан"},{"name":"BIC","type":"STRING","value":{"asString":"049514608","typeOfValue":"STRING","value":"049514608"},"valueAsOfType":"049514608"},{"name":"CORR_ACCOUNT","type":"STRING","value":{"asString":"30101810500000000608","typeOfValue":"STRING","value":"30101810500000000608"},"valueAsOfType":"30101810500000000608"}],"attributeValues":{"ID":"049514608","CORR_ACCOUNT":"30101810500000000608","BIC":"049514608","NAME":"АБАКАНСКОЕ ОТДЕЛЕНИЕ N8602 ПАО СБЕРБАНК г Абакан"}},{"value":"041012765","title":"041012765 - \"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск","isLeaf":true,"children":[],"attributes":[{"name":"ID","type":"STRING","value":{"asString":"041012765","typeOfValue":"STRING","value":"041012765"},"valueAsOfType":"041012765"},{"name":"NAME","type":"STRING","value":{"asString":"\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск","typeOfValue":"STRING","value":"\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск"},"valueAsOfType":"\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск"},{"name":"BIC","type":"STRING","value":{"asString":"041012765","typeOfValue":"STRING","value":"041012765"},"valueAsOfType":"041012765"},{"name":"CORR_ACCOUNT","type":"STRING","value":{"asString":"30101810300000000765","typeOfValue":"STRING","value":"30101810300000000765"},"valueAsOfType":"30101810300000000765"}],"attributeValues":{"ID":"041012765","CORR_ACCOUNT":"30101810300000000765","BIC":"041012765","NAME":"\"Азиатско-Тихоокеанский Банк\" (АО) г Благовещенск"}}]`
	dictSuccessEmptyResponse   = `{"error":{"code":0,"message":"operation completed"},"fieldErrors":[],"total":5004,"items":[]}`
	dictErrorResponse          = `{"error":{"code":7,"message":"Entity not found"},"fieldErrors":[],"total":0,"items":[]}`
)

type pushV2Method struct {
	name   string
	path   string
	parent error
	call   func(*Client, context.Context, *Archive) (*OrderResult, error)
}

func pushV2Methods() []pushV2Method {
	return []pushV2Method{
		{"OrderPushV2", "/api/gusmev/v2/push", ErrPush, func(c *Client, ctx context.Context, a *Archive) (*OrderResult, error) {
			return c.OrderPushV2(ctx, testToken, testMeta, a)
		}},
		{"OrderPushChunkedV2", "/api/gusmev/v2/push/chunked", ErrPushChunked, func(c *Client, ctx context.Context, a *Archive) (*OrderResult, error) {
			return c.OrderPushChunkedV2(ctx, testToken, 123456, a)
		}},
	}
}

func (suite *suiteTestClient) TestPushV2Requests() {
	for _, method := range pushV2Methods() {
		for _, size := range []int{1, 3, 4, 5, 8, 9} {
			for _, name := range []string{"test-archive", ""} {
				suite.Run(fmt.Sprintf("%s/size=%d/name=%s", method.name, size, name), func() {
					data := bytes.Repeat([]byte("x"), size)
					var received []byte
					count := 0
					total := 1
					if method.parent == ErrPushChunked {
						total = (size + 3) / 4
					}
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						current := count
						count++
						suite.Equal(http.MethodPost, r.Method)
						suite.Equal(method.path, r.URL.Path)
						suite.Equal("Bearer "+testToken, r.Header.Get("Authorization"))
						suite.Contains(r.Header.Get("Content-Type"), "multipart/form-data; boundary=")
						if !suite.NoError(r.ParseMultipartForm(1024)) {
							return
						}
						defer r.MultipartForm.RemoveAll()
						if method.parent == ErrPush {
							suite.JSONEq(string(testMeta.JSON()), r.FormValue("meta"))
							suite.NotContains(r.Form, "orderId")
						} else {
							suite.NotContains(r.Form, "meta")
							suite.Equal("123456", r.FormValue("orderId"))
						}
						if total > 1 {
							suite.Equal(strconv.Itoa(current), r.FormValue("chunk"))
							suite.Equal(strconv.Itoa(total), r.FormValue("chunks"))
						} else {
							suite.NotContains(r.Form, "chunk")
							suite.NotContains(r.Form, "chunks")
						}
						if !suite.Len(r.MultipartForm.File["file"], 1) {
							return
						}
						fh := r.MultipartForm.File["file"][0]
						filename := name
						if filename == "" {
							filename = DefaultArchiveName
						}
						extension := ".zip"
						if total > 1 {
							extension = fmt.Sprintf(".z%03d", current+1)
						}
						suite.Equal(filename+extension, fh.Filename)
						suite.Equal("application/octet-stream", fh.Header.Get("Content-Type"))
						f, err := fh.Open()
						if !suite.NoError(err) {
							return
						}
						part, err := io.ReadAll(f)
						f.Close()
						suite.NoError(err)
						if total > 1 && current < total-1 {
							suite.Len(part, 4)
						}
						received = append(received, part...)
						w.Header().Set("Content-Type", "application/json")
						if current < total-1 {
							w.WriteHeader(http.StatusPartialContent)
						}
						fmt.Fprint(w, `{"orderId":123456}`)
					}))
					defer server.Close()
					result, err := method.call(NewClient(server.URL).WithChunkSize(4), context.Background(), &Archive{Name: name, Data: data})
					suite.NoError(err)
					suite.Equal(&OrderResult{OrderId: 123456}, result)
					suite.Equal(total, count)
					suite.Equal(data, received)
				})
			}
		}
	}
}

func (suite *suiteTestClient) TestPushV2Messages() {
	for _, method := range pushV2Methods() {
		for _, message := range []string{"", `,"message":""`, `,"message":null`, `,"message":"Ведомственные работы. Сервис ведомства работает с перебоями"`} {
			suite.Run(method.name+message, func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, `{"orderId":123456%s}`, message) }))
				defer server.Close()
				result, err := method.call(NewClient(server.URL), context.Background(), &Archive{Data: []byte("x")})
				suite.NoError(err)
				expected := ""
				if message == `,"message":"Ведомственные работы. Сервис ведомства работает с перебоями"` {
					expected = "Ведомственные работы. Сервис ведомства работает с перебоями"
				}
				suite.Equal(&OrderResult{OrderId: 123456, Message: expected}, result)
			})
		}
	}
	for _, tc := range []struct {
		name     string
		messages []string
		want     string
	}{
		{"none", []string{"", `,"message":""`, `,"message":null`}, ""},
		{"first survives empty and absent", []string{`,"message":"первое"`, `,"message":""`, ""}, "первое"},
		{"first survives null", []string{`,"message":"первое"`, `,"message":null`}, "первое"},
		{"last only", []string{"", `,"message":"последнее"`}, "последнее"},
		{"last nonempty", []string{`,"message":"первое"`, `,"message":"второе"`, `,"message":""`}, "второе"},
		{"repeated", []string{`,"message":"работы"`, `,"message":"работы"`}, "работы"},
		{"text preserved", []string{`,"message":"  работы\n"`, ""}, "  работы\n"},
	} {
		suite.Run(tc.name, func() {
			count := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				message := tc.messages[count]
				count++
				if count < len(tc.messages) {
					w.WriteHeader(http.StatusPartialContent)
				}
				fmt.Fprintf(w, `{"orderId":123456%s}`, message)
			}))
			defer server.Close()
			result, err := NewClient(server.URL).WithChunkSize(1).OrderPushChunkedV2(context.Background(), testToken, 123456, &Archive{Data: bytes.Repeat([]byte("x"), len(tc.messages))})
			suite.NoError(err)
			suite.Equal(&OrderResult{OrderId: 123456, Message: tc.want}, result)
			suite.Equal(len(tc.messages), count)
		})
	}
}

func (suite *suiteTestClient) TestPushV2Errors() {
	for _, method := range pushV2Methods() {
		for _, tc := range []struct {
			name              string
			status            int
			body, contentType string
			want              []error
		}{
			{"missing ID", 200, `{}`, "application/json", []error{ErrWrongOrderID}},
			{"zero ID", 200, `{"orderId":0}`, "application/json", []error{ErrWrongOrderID}},
			{"malformed JSON", 200, `{`, "application/json", []error{ErrJSONUnmarshal}},
			{"empty body", 200, ``, "application/json", []error{ErrJSONUnmarshal}},
			{"wrong ID type", 200, `{"orderId":"123456"}`, "application/json", []error{ErrJSONUnmarshal}},
			{"wrong message type", 200, `{"orderId":123456,"message":123}`, "application/json", []error{ErrJSONUnmarshal}},
			{"service unavailable", 500, `{"code":"service_unavailable","message":"Ведомственная авария. Попробуйте подать заявление позже"}`, "application/json", []error{ErrStatusInternalError, ErrCodeServiceUnavailable}},
			{"HTTP 503", 503, ``, "text/plain", []error{ErrStatusServiceUnavailable}},
			{"known code", 403, `{"code":"access_denied_service","message":"нет доступа"}`, "application/json", []error{ErrStatusForbidden, ErrCodeAccessDeniedService}},
			{"unknown code", 400, `{"code":"new_code","message":"ошибка"}`, "application/json", []error{ErrStatusBadRequest, ErrCodeUnexpected}},
			{"malformed error", 400, `{`, "application/json", []error{ErrStatusBadRequest, ErrJSONUnmarshal}},
			{"plain text error", 500, `error`, "text/plain", []error{ErrStatusInternalError, ErrUnexpectedContentType}},
		} {
			suite.Run(method.name+"/"+tc.name, func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", tc.contentType)
					w.WriteHeader(tc.status)
					fmt.Fprint(w, tc.body)
				}))
				defer server.Close()
				result, err := method.call(NewClient(server.URL), context.Background(), &Archive{Data: []byte("x")})
				suite.Nil(result)
				suite.ErrorIs(err, method.parent)
				for _, want := range tc.want {
					suite.ErrorIs(err, want)
				}
				if errors.Is(err, ErrCodeServiceUnavailable) {
					suite.Contains(err.Error(), "Ведомственная авария. Попробуйте подать заявление позже")
					suite.NotErrorIs(err, ErrStatusServiceUnavailable)
				}
			})
		}
		for i, archive := range []*Archive{nil, {}, {Data: []byte{}}} {
			suite.Run(fmt.Sprintf("%s/empty archive %d", method.name, i), func() {
				result, err := method.call(NewClient(""), context.Background(), archive)
				suite.Nil(result)
				suite.ErrorIs(err, method.parent)
				suite.ErrorIs(err, ErrNilArchive)
			})
		}
		suite.Run(method.name+"/transport", func() {
			result, err := method.call(NewClient(":"), context.Background(), &Archive{Data: []byte("x")})
			suite.Nil(result)
			suite.ErrorIs(err, method.parent)
			suite.ErrorIs(err, ErrRequest)
		})
		suite.Run(method.name+"/canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			result, err := method.call(NewClient("http://127.0.0.1"), ctx, &Archive{Data: []byte("x")})
			suite.Nil(result)
			suite.ErrorIs(err, method.parent)
			suite.ErrorIs(err, ErrRequest)
			suite.ErrorIs(err, context.Canceled)
		})
	}
}

func (suite *suiteTestClient) TestPushChunkedV2StopsOnError() {
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"wrong ID", `{"orderId":999}`, 200, ErrWrongOrderID},
		{"missing ID", `{}`, 200, ErrWrongOrderID},
		{"invalid JSON", `{`, 200, ErrJSONUnmarshal},
		{"blocked", `{"code":"service_unavailable","message":"авария"}`, 500, ErrCodeServiceUnavailable},
		{"canceled", `{"orderId":123456}`, 206, context.Canceled},
	} {
		suite.Run(tc.name, func() {
			count := 0
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				w.Header().Set("Content-Type", "application/json")
				if count == 1 {
					w.WriteHeader(206)
					fmt.Fprint(w, `{"orderId":123456,"message":"работы"}`)
					return
				}
				if tc.name == "canceled" {
					cancel()
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			result, err := NewClient(server.URL).WithChunkSize(1).OrderPushChunkedV2(ctx, testToken, 123456, &Archive{Data: []byte("abc")})
			suite.Nil(result)
			suite.ErrorIs(err, ErrPushChunked)
			suite.ErrorIs(err, tc.want)
			suite.Equal(2, count)
		})
	}
}

// Старые методы игнорируют message, в том числе неизвестный им формат поля.
func (suite *suiteTestClient) TestPushLegacyIgnoresMessage() {
	for _, message := range []string{`"работы"`, `123`, `{}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.NotContains(r.URL.Path, "/v2/")
			fmt.Fprintf(w, `{"orderId":123456,"message":%s}`, message)
		}))
		client := NewClient(server.URL)
		archive := &Archive{Data: []byte("x")}
		orderId, err := client.OrderPush(context.Background(), testToken, testMeta, archive)
		suite.NoError(err)
		suite.Equal(123456, orderId)
		suite.NoError(client.OrderPushChunked(context.Background(), testToken, 123456, archive))
		server.Close()
	}
}

type pushV2RoundTripper func(*http.Request) (*http.Response, error)

func (f pushV2RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type pushV2ErrorReader struct{ err error }

func (r pushV2ErrorReader) Read([]byte) (int, error) { return 0, r.err }

func (suite *suiteTestClient) TestPushV2IOErrors() {
	failure := errors.New("test network failure")
	for _, method := range pushV2Methods() {
		for _, tc := range []struct {
			name      string
			transport pushV2RoundTripper
		}{
			{"round trip", func(r *http.Request) (*http.Response, error) { return nil, failure }},
			{"read body", func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(pushV2ErrorReader{failure})}, nil
			}},
		} {
			suite.Run(method.name+"/"+tc.name, func() {
				client := NewClient("http://example.test").WithHTTPClient(&http.Client{Transport: tc.transport})
				result, err := method.call(client, context.Background(), &Archive{Data: []byte("x")})
				suite.Nil(result)
				suite.ErrorIs(err, method.parent)
				suite.ErrorIs(err, ErrRequest)
				suite.ErrorIs(err, failure)
			})
		}
		suite.Run(method.name+"/deadline", func() {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			result, err := method.call(NewClient("http://127.0.0.1"), ctx, &Archive{Data: []byte("x")})
			suite.Nil(result)
			suite.ErrorIs(err, method.parent)
			suite.ErrorIs(err, ErrRequest)
			suite.ErrorIs(err, context.DeadlineExceeded)
		})
	}
}

// Полный ответ из раздела 4.5 спецификации API ЕПГУ версии 1.15.
// Фикстура сохранена в тестах, чтение DOCX при запуске не требуется.
const orderInfoSpecificationV115Response = `{"code":"OK","message":null,"message_id":"71e35026-ade1-11e9-9931-00155d040494","order":"{\"orderType\":\"ORDER\",\"smevTx\":\"b0ba0139-613d-4d4d-bc28-4240efc9a6ec\",\"hasEmpowerment2021\":false,\"smevMessageId\":\"WAIT_RESPONSE\",\"formVersion\":\"1\",\"ownerId\":1078941357,\"notifySms\":\"\",\"hasTimestamp\":false,\"orderStatusName\":\"Ошибка обработки заявления\",\"portalName\":\"Портал гос услуг v1\",\"paymentRequired\":false,\"id\":2058864134,\"signCnt\":0,\"childrenSigned\":false,\"stateOrgStatusCode\":\"\",\"inviteToEqueueUrl\":\"\",\"statusColorCode\":\"\",\"orderStatusId\":24,\"withCustomResult\":true,\"hasNoPaidPayment\":false,\"servicePassportId\":\"600113\",\"routingCode\":\"\",\"sourceSystem\":\"PGU\",\"stateOrgStatusName\":\"\",\"serviceUrl\":\"\",\"serviceEpguId\":\"1\",\"edsStatus\":\"EDS_NOT_SUPPORTED\",\"allowToDelete\":true,\"requestDate\":\"2022-07-25T13:01:47.438+0300\",\"complexMode\":\"\",\"stateOrgCode\":\"pfr\",\"personType\":\"PERSON\",\"serviceTargetId\":\"-10000000113\",\"orderPayments\":[],\"orderResponseFiles\":[],\"hasResult\":false,\"userId\":1078941357,\"allowToEdit\":false,\"closed\":true,\"readyToSign\":false,\"currentStatusHistory\":{\"date\":\"2022-07-25T13:01:52.121+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":true,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Ошибка обработки заявления\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":24,\"editAllowed\":false,\"sender\":\"Пенсионный фонд Российской Федерации\",\"mnemonic\":\"\",\"comment\":\"\",\"id\":6268157772},\"infoMessages\":[],\"notifyEmail\":\"\",\"paymentCount\":0,\"draftHidden\":false,\"stateStructureId\":\"10000002796\",\"hasNewStatus\":true,\"stateOrgId\":266,\"redirectUrl\":\"\",\"hasChildren\":false,\"stateStructureName\":\"ПФР\",\"portalCode\":\"PGU\",\"esepOperationId\":\"\",\"orderAttributeEvents\":[],\"notifyPush\":\"\",\"eserviceId\":\"10000000113\",\"currentStatusHistoryId\":6268157772,\"linkToOrderForm\":\"\",\"extOrderUrl\":\"\",\"paymentStatusEvents\":[],\"payback\":false,\"readyToPush\":false,\"admLevelCode\":\"FEDERAL\",\"formPrefilling\":false,\"allFileSign\":false,\"orgUserName\":\"\",\"noPaidPaymentCount\":-1,\"creationMode\":\"direct\",\"steps\":[],\"powerMnemonic\":\"\",\"statuses\":[{\"date\":\"2022-07-25T13:01:46.413+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Черновик заявления\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":0,\"editAllowed\":false,\"sender\":\"\",\"mnemonic\":\"\",\"comment\":\"\",\"id\":6268157768},{\"date\":\"2022-07-25T13:01:47.426+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Зарегистрировано на портале\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":17,\"editAllowed\":false,\"sender\":\"Пенсионный фонд Российской Федерации\",\"mnemonic\":\"\",\"comment\":\"\",\"id\":6268157769},{\"date\":\"2022-07-25T13:01:48.520+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Заявление отправлено в ведомство\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":21,\"editAllowed\":false,\"sender\":\"Пенсионный фонд Российской Федерации\",\"mnemonic\":\"\",\"comment\":\"\",\"id\":6268157770},{\"date\":\"2022-07-25T13:01:51.810+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":false,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Заявление получено ведомством\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":2,\"editAllowed\":false,\"sender\":\"Пенсионный фонд Российской Федерации\",\"mnemonic\":\"\",\"comment\":\"Сообщение доставлено\",\"id\":6268157771},{\"date\":\"2022-07-25T13:01:52.121+0300\",\"cancelAllowed\":false,\"unreadEvent\":true,\"deliveryCancelAllowed\":false,\"finalStatus\":true,\"orderId\":2058864134,\"stateOrgStatusCode\":\"\",\"author\":\"\",\"hasResult\":\"N\",\"stateOrgStatusDescr\":\"\",\"title\":\"Ошибка обработки заявления\",\"mfcFinalStatus\":false,\"sendMessageAllowed\":false,\"statusId\":24,\"editAllowed\":false,\"sender\":\"Пенсионный фонд Российской Федерации\",\"mnemonic\":\"\",\"comment\":\"\",\"id\":6268157772}],\"orderDate\":\"2022-07-25T13:01:46.000+0300\",\"updated\":\"2022-07-25T13:01:52.129+0300\",\"checkQueue\":false,\"withDelivery\":false,\"gisdo\":false,\"userSelectedRegion\":\"00000000000\",\"description\":\"\",\"eQueueEvents\":[],\"hasActiveInviteToEqueue\":false,\"multRegion\":false,\"extSystem\":false,\"useAsTemplate\":false,\"qrlink\":{\"hasAltMimeType\":false,\"fileName\":\"\",\"edsStatus\":\"\",\"fileSize\":0,\"canSentToMFC\":false,\"link\":\"\",\"id\":\"\",\"mimeType\":\"\",\"hasDigitalSignature\":false,\"additionalName\":\"\"},\"hasPreviewPdf\":false,\"testUser\":false,\"textMessages\":[],\"unreadMessageCnt\":0,\"parentOrderStateStructureName\":\"\",\"cr_uin\":\"\",\"infSysCode\":\"\",\"serviceName\":\"Выдача справки о размере пенсии и иных социальных выплат на дату\",\"deprecatedService\":false,\"hubForm\":false,\"extOrderId\":\"\",\"orderAttachmentFiles\":[{\"fileName\":\"req.xml\",\"fileSize\":653,\"link\":\"terrabyte://00/2058864134/req.xml/2\",\"id\":\"2058864134/files/cmVxLnhtbA\",\"mimeType\":\"application/xml\",\"hasDigitalSignature\":false,\"type\":\"REQUEST\"}],\"online\":false,\"location\":\"36\"}"}`

func (suite *suiteTestClient) TestOrderInfoSpecificationV115() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderInfoSpecificationV115Response)
	}))
	defer server.Close()
	result, err := NewClient(server.URL).OrderInfo(context.Background(), testToken, 2058864134)
	suite.Require().NoError(err)
	suite.Equal("OK", result.Code)
	suite.Equal("71e35026-ade1-11e9-9931-00155d040494", result.MessageId)
	suite.Require().NotNil(result.Order)
	suite.Equal(2058864134, result.Order.Id)
	suite.Equal(24, result.Order.OrderStatusId)
	suite.Len(result.Order.Statuses, 5)
	suite.Equal("Ошибка обработки заявления", result.Order.CurrentStatusHistory.Title)
	suite.Len(result.Order.OrderAttachmentFiles, 1)
	suite.Equal("req.xml", result.Order.OrderAttachmentFiles[0].FileName)
	suite.False(result.Order.OrderDate.IsZero())
	suite.False(result.Order.CurrentStatusHistory.Date.IsZero())
}

func (suite *suiteTestClient) TestOrderInfoAdditionalFields() {
	// В документе эти поля пусты. Непустые значения проверяют их разбор и JSON-теги.
	order := `{
  "inviteToEqueueUrl":"invite", "statusColorCode":"color", "complexMode":"complex",
  "redirectUrl":"redirect", "esepOperationId":"operation", "linkToOrderForm":"form",
  "extOrderUrl":"external", "description":"description", "parentOrderStateStructureName":"parent",
  "cr_uin":"uin", "infSysCode":"system", "extOrderId":"external-id", "formPrefilling":true,
  "statuses":[{"author":"status-author", "mfcFinalStatus":true}],
  "currentStatusHistory":{"author":"history-author", "mfcFinalStatus":true, "mnemonic":"sender"},
  "qrlink":{"fileName":"qr.pdf", "edsStatus":"signed", "link":"storage-link",
   "id":"file-id", "mimeType":"application/pdf", "additionalName":"additional"}
 }`
	body, err := json.Marshal(map[string]any{"code": "OK", "order": order})
	suite.Require().NoError(err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer server.Close()
	result, err := NewClient(server.URL).OrderInfo(context.Background(), testToken, 123456)
	suite.Require().NoError(err)
	suite.Require().NotNil(result.Order)
	details := result.Order
	suite.Equal("invite", details.InviteToEqueueUrl)
	suite.Equal("color", details.StatusColorCode)
	suite.Equal("complex", details.ComplexMode)
	suite.Equal("redirect", details.RedirectUrl)
	suite.Equal("operation", details.EsepOperationId)
	suite.Equal("form", details.LinkToOrderForm)
	suite.Equal("external", details.ExtOrderUrl)
	suite.Equal("description", details.Description)
	suite.Equal("parent", details.ParentOrderStateStructureName)
	suite.Equal("uin", details.CrUin)
	suite.Equal("system", details.InfSysCode)
	suite.Equal("external-id", details.ExtOrderId)
	suite.True(details.FormPrefilling)
	suite.Require().Len(details.Statuses, 1)
	suite.Equal("status-author", details.Statuses[0].Author)
	suite.True(details.Statuses[0].MFCFinalStatus)
	suite.Equal("history-author", details.CurrentStatusHistory.Author)
	suite.True(details.CurrentStatusHistory.MFCFinalStatus)
	suite.Equal("sender", details.CurrentStatusHistory.Mnemonic)
	suite.Equal("qr.pdf", details.Qrlink.FileName)
	suite.Equal("signed", details.Qrlink.EdsStatus)
	suite.Equal("storage-link", details.Qrlink.Link)
	suite.Equal("file-id", details.Qrlink.Id)
	suite.Equal("application/pdf", details.Qrlink.MimeType)
	suite.Equal("additional", details.Qrlink.AdditionalName)
}

func (suite *suiteTestClient) TestOrderInfoMessageIdVariants() {
	for _, tc := range []struct{ name, fields, want string }{
		{"table spelling", `"messageId":"canonical"`, "canonical"},
		{"example spelling", `"message_id":"example"`, "example"},
		{"both", `"messageId":"canonical", "message_id":"example"`, "canonical"},
		{"both reversed", `"message_id":"example", "messageId":"canonical"`, "canonical"},
		{"empty canonical", `"messageId":"", "message_id":"example"`, "example"},
		{"null canonical", `"messageId":null, "message_id":"example"`, "example"},
		{"both empty", `"messageId":"", "message_id":""`, ""},
		{"absent", `"message":null`, ""},
	} {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"code":"OK",%s,"order":null}`, tc.fields)
			}))
			defer server.Close()
			result, err := NewClient(server.URL).OrderInfo(context.Background(), testToken, 123456)
			suite.Require().NoError(err)
			suite.Equal(tc.want, result.MessageId)
			suite.Nil(result.Order)
		})
	}
}
