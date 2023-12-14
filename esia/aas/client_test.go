package aas

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-api-epgu/esia/signature"
	"github.com/ofstudio/go-api-epgu/utils"
)

const (
	testSignature = "this is a test signature"
	testCertHash  = "test_hash"
)

type suiteTestClient struct {
	suite.Suite
}

func TestClient(t *testing.T) {
	suite.Run(t, new(suiteTestClient))
}

func (suite *suiteTestClient) TearDownSubTest() {
	guid = utils.GUID
}

func (suite *suiteTestClient) TestAuthURI() {

	suite.Run("success", func() {
		guid = func() (string, error) {
			return "test-state", nil
		}

		client := NewClient("", "test-client", signature.NewNop(testSignature, testCertHash))
		var permissions = Permissions{
			{
				ResponsibleObject: "test",
				Sysname:           "test",
				Expire:            1,
				Actions:           []PermissionAction{{Sysname: "test"}},
				Purposes:          []PermissionPurpose{{Sysname: "test"}},
				Scopes:            []PermissionScope{{Sysname: "test"}},
			},
		}

		uriStr, err := client.AuthURI("test-scope", "test-redirect", permissions)
		suite.NoError(err)
		u, err := url.Parse(uriStr)
		suite.NoError(err)

		q := u.Query()
		suite.Equal(UserEndpoint, u.Path)
		suite.Equal("test-client", q.Get("client_id"))
		suite.Equal("test-scope", q.Get("scope"))
		suite.Equal("test-state", q.Get("state"))
		suite.Equal("test-redirect", q.Get("redirect_uri"))
		suite.Equal("code", q.Get("response_type"))
		suite.Equal(permissions.Base64String(), q.Get("permissions"))
		suite.Equal(testCertHash, q.Get("client_certificate_hash"))
		suite.Regexp(`^\d{4}.\d{2}.\d{2} \d{2}:\d{2}:\d{2} [\+-]\d{4}$`, q.Get("timestamp")) // timestamp

		sign, err := base64.URLEncoding.DecodeString(q.Get("client_secret"))
		suite.NoError(err)
		suite.Equal(testSignature, string(sign))
		suite.Equal("online", q.Get("access_type"))
	})

	suite.Run("error guid", func() {
		client := NewClient("", "test-client", signature.NewNop(testSignature, testCertHash))
		guid = func() (string, error) {
			return "", ErrGUID
		}
		uriStr, err := client.AuthURI("test-scope", "test-redirect", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrGUID)
		suite.Empty(uriStr)
	})

	suite.Run("error sign", func() {
		client := NewClient("", "test", signature.NewNop("", ""))
		uriStr, err := client.AuthURI("openid", "test", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrSign)
		suite.Empty(uriStr)
	})

	suite.Run("error signer is nil", func() {
		client := NewClient("", "test", nil)
		uriStr, err := client.AuthURI("openid", "test", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrSign)
		suite.Empty(uriStr)
	})

}

func (suite *suiteTestClient) TestParseCallback() {
	suite.Run("success", func() {
		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		code, state, err := client.ParseCallback(url.Values{
			"code":  []string{"test-code"},
			"state": []string{"test-state"},
		})
		suite.NoError(err)
		suite.Equal("test-code", code)
		suite.Equal("test-state", state)
	})

	suite.Run("error no state", func() {
		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		code, state, err := client.ParseCallback(url.Values{
			"code": []string{"ESIA-007014: The request doesn't contain..."},
		})
		suite.ErrorIs(err, ErrParseCallback)
		suite.ErrorIs(err, ErrNoState)
		suite.Equal("ошибка обратного вызова: отсутствует поле state", err.Error())
		suite.Empty(code)
		suite.Empty(state)
	})

	suite.Run("error ESIA", func() {
		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		code, state, err := client.ParseCallback(url.Values{
			"state":             []string{"test"},
			"error":             []string{"invalid_request"},
			"error_description": []string{"ESIA-007014: The request doesn't contain..."},
		})
		suite.ErrorIs(err, ErrParseCallback)
		suite.ErrorIs(err, ErrESIA_007014)
		suite.Equal(
			"ошибка обратного вызова: ESIA-007014: Запрос не содержит обязательного параметра [error='invalid_request', error_description='ESIA-007014: The request doesn't contain...', state='test']",
			err.Error(),
		)
		suite.Empty(code)
		suite.Equal("test", state)
	})

	suite.Run("error ESIA unknown", func() {
		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		code, state, err := client.ParseCallback(url.Values{
			"state":             []string{"test"},
			"error":             []string{"invalid_request"},
			"error_description": []string{"ESIA-999999: The request doesn't contain..."},
		})

		suite.ErrorIs(err, ErrParseCallback)
		suite.ErrorIs(err, ErrESIA_unknown)
		suite.Equal(
			"ошибка обратного вызова: неизвестная ошибка ЕСИА [error='invalid_request', error_description='ESIA-999999: The request doesn't contain...', state='test']",
			err.Error(),
		)
		suite.Empty(code)
		suite.Equal("test", state)
	})
}

func (suite *suiteTestClient) TestTokenExchange() {
	suite.Run("success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/aas/oauth2/v3/te", r.URL.Path)
			suite.Equal("application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			suite.Equal("test", r.FormValue("client_id"))
			suite.Equal(base64.URLEncoding.EncodeToString([]byte(testSignature)), r.FormValue("client_secret"))
			suite.Equal("test-code", r.FormValue("code"))
			suite.Equal("test-scope", r.FormValue("scope"))
			suite.Equal(testCertHash, r.FormValue("client_certificate_hash"))
			suite.Equal("test-uri", r.FormValue("redirect_uri"))
			suite.Equal("authorization_code", r.FormValue("grant_type"))
			suite.Equal("Bearer", r.FormValue("token_type"))
			suite.Regexp(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`, r.FormValue("state")) // guid
			suite.Regexp(`^\d{4}.\d{2}.\d{2} \d{2}:\d{2}:\d{2} [\+-]\d{4}$`, r.FormValue("timestamp"))                 // timestamp

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"test","id_token":"test","state":"test","token_type":"Bearer","expires_in":0}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test-code", "test-scope", "test-uri")
		suite.NoError(err)
		suite.Require().NotNil(token)
		suite.Equal("test", token.AccessToken)
	})

	suite.Run("error 500 unexpected content type", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("test"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrUnexpectedContentType)
		suite.Equal(
			"ошибка запроса токена: HTTP 500 Internal Server Error: неожиданный тип содержимого: 'text/html'",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error 400 ESIA-007004", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"access_denied","error_description":"ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос","state":"test"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrESIA_007004)
		suite.Equal(
			"ошибка запроса токена: HTTP 400 Bad Request: ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос [error='access_denied', error_description='ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос', state='test']",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error 200 malformed json", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not a json"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(
			"ошибка запроса токена: ошибка чтения JSON: invalid character 'o' in literal null (expecting 'u')",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error 400 malformed json", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("not a json"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(
			"ошибка запроса токена: HTTP 400 Bad Request: ошибка чтения JSON: invalid character 'o' in literal null (expecting 'u')",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error guid", func() {
		guid = func() (string, error) {
			return "", errors.New("test")
		}

		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrGUID)
		suite.Nil(token)
	})

	suite.Run("error request call", func() {
		client := NewClient("", "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrRequest)
		suite.Nil(token)
	})

	suite.Run("error sign", func() {
		client := NewClient("", "test", signature.NewNop("", ""))
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

	suite.Run("error signer is nil", func() {
		client := NewClient("", "test", nil)
		token, err := client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})
}

func (suite *suiteTestClient) TestTokenUpdate() {

	suite.Run("success", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.Equal(http.MethodPost, r.Method)
			suite.Equal("/aas/oauth2/v3/te", r.URL.Path)
			suite.Equal("application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			suite.Equal("test-client", r.FormValue("client_id"))
			suite.Equal(base64.URLEncoding.EncodeToString([]byte(testSignature)), r.FormValue("client_secret"))
			suite.Equal("prm_chg?oid=test-oid", r.FormValue("scope"))
			suite.Equal(testCertHash, r.FormValue("client_certificate_hash"))
			suite.Equal("test-redirect", r.FormValue("redirect_uri"))
			suite.Equal("client_credentials", r.FormValue("grant_type"))
			suite.Equal("Bearer", r.FormValue("token_type"))
			suite.Regexp(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`, r.FormValue("state")) // guid
			suite.Regexp(`^\d{4}.\d{2}.\d{2} \d{2}:\d{2}:\d{2} [\+-]\d{4}$`, r.FormValue("timestamp"))                 // timestamp

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"test","id_token":"test","state":"test","token_type":"Bearer","expires_in":0}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-client", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.NoError(err)
		suite.Require().NotNil(token)
		suite.Equal("test", token.AccessToken)
	})

	suite.Run("error 500 unexpected content type", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("test"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrUnexpectedContentType)
		suite.Nil(token)
	})

	suite.Run("error 400 ESIA-007004", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"access_denied","error_description":"ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос","state":"test"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-client", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrESIA_007004)
		suite.Equal(
			"ошибка обновления токена: HTTP 400 Bad Request: ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос [error='access_denied', error_description='ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос', state='test']",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error malformed json", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not a json"))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-client", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrJSONUnmarshal)
		suite.Equal(
			"ошибка обновления токена: ошибка чтения JSON: invalid character 'o' in literal null (expecting 'u')",
			err.Error(),
		)
		suite.Nil(token)
	})

	suite.Run("error guid", func() {
		guid = func() (string, error) {
			return "", errors.New("test")
		}

		client := NewClient("", "test-client", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrGUID)
		suite.Nil(token)
	})

	suite.Run("error request call", func() {
		client := NewClient("", "test-client", signature.NewNop(testSignature, testCertHash))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrRequest)
		suite.Nil(token)
	})

	suite.Run("error sign", func() {
		client := NewClient("", "test-client", signature.NewNop("", ""))
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

	suite.Run("error signer is nil", func() {
		client := NewClient("", "test-client", nil)
		token, err := client.TokenUpdate("test-oid", "test-redirect")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})
}
