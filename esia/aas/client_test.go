package aas

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-api-epgu/esia/signature"
	"github.com/ofstudio/go-api-epgu/utils"
)

type suiteTestClient struct {
	suite.Suite
	client *Client
}

func TestClient(t *testing.T) {
	suite.Run(t, new(suiteTestClient))
}
func (suite *suiteTestClient) SetupTest() {
	suite.client = NewClient("http://test.gock", "test", signature.NewNop(testSignature, testCertHash))
}

func (suite *suiteTestClient) TearDownSubTest() {
	guid = utils.GUID
	gock.Off()
}

func (suite *suiteTestClient) TestAuthURI() {

	suite.Run("success", func() {
		guid = func() (string, error) {
			return "test", nil
		}

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

		uriStr, err := suite.client.AuthURI("openid", "test", permissions)
		suite.NoError(err)
		u, err := url.Parse(uriStr)
		suite.NoError(err)

		q := u.Query()
		suite.Equal(UserEndpoint, u.Path)
		suite.Equal("test", q.Get("client_id"))
		suite.Equal("openid", q.Get("scope"))
		suite.Equal("test", q.Get("state"))
		suite.Equal("test", q.Get("redirect_uri"))
		suite.Equal("code", q.Get("response_type"))
		suite.Equal(permissions.Base64String(), q.Get("permissions"))
		suite.Equal(testCertHash, q.Get("client_certificate_hash"))
		suite.NotEmpty(q.Get("timestamp"))

		sign, err := base64.URLEncoding.DecodeString(q.Get("client_secret"))
		suite.NoError(err)
		suite.Equal(testSignature, string(sign))
		suite.Equal("online", q.Get("access_type"))
	})

	suite.Run("error guid", func() {
		guid = func() (string, error) {
			return "", ErrGUID
		}
		uriStr, err := suite.client.AuthURI("openid", "test", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrGUID)
		suite.Empty(uriStr)
	})

	suite.Run("error sign", func() {
		suite.client.signer = signature.NewNop("", "")
		uriStr, err := suite.client.AuthURI("openid", "test", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrSign)
		suite.Empty(uriStr)
	})

	suite.Run("error signer is nil", func() {
		suite.client.signer = nil
		uriStr, err := suite.client.AuthURI("openid", "test", Permissions{})
		suite.ErrorIs(err, ErrAuthURI)
		suite.ErrorIs(err, ErrSign)
		suite.Empty(uriStr)
	})

}

func (suite *suiteTestClient) TestParseCallback() {
	suite.Run("success", func() {
		code, state, err := suite.client.ParseCallback(url.Values{
			"code":  []string{"test"},
			"state": []string{"test"},
		})
		suite.NoError(err)
		suite.Equal("test", code)
		suite.Equal("test", state)
	})

	suite.Run("error no state", func() {
		code, state, err := suite.client.ParseCallback(url.Values{
			"code": []string{"test"},
		})
		suite.ErrorIs(err, ErrParseCallback)
		suite.ErrorIs(err, ErrNoState)
		suite.Empty(code)
		suite.Empty(state)
	})

	suite.Run("error ESIA", func() {
		code, state, err := suite.client.ParseCallback(url.Values{
			"state":             []string{"test"},
			"error":             []string{"invalid_request"},
			"error_description": []string{"ESIA-007014: The request doesn't contain the mandatory parameter [client_certificate_hash]."},
		})
		suite.ErrorIs(err, ErrParseCallback)
		suite.ErrorIs(err, ErrESIA_007014)
		suite.Empty(code)
		suite.Equal("test", state)
	})
}

func (suite *suiteTestClient) TestTokenExchange() {
	suite.Run("success", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			MatchType("application/x-www-form-urlencoded").
			AddMatcher(matchFormField("client_id", "test")).
			AddMatcher(matchFormField("client_secret", base64.URLEncoding.EncodeToString([]byte(testSignature)))).
			AddMatcher(matchFormField("code", "test")).
			AddMatcher(matchFormField("scope", "test")).
			AddMatcher(matchFormField("client_certificate_hash", testCertHash)).
			AddMatcher(matchFormField("redirect_uri", "test")).
			AddMatcher(matchFormField("grant_type", "authorization_code")).
			AddMatcher(matchFormField("token_type", "Bearer")).
			AddMatcher(matchFormFieldRegexp("state", `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)). // guid
			AddMatcher(matchFormFieldRegexp("timestamp", `^\d{4}.\d{2}.\d{2} \d{2}:\d{2}:\d{2} [\+-]\d{4}$`)).
			Reply(200).
			JSON(TokenExchangeResponse{
				AccessToken: "test",
				IdToken:     "test",
				State:       "test",
				TokenType:   "Test",
				ExpiresIn:   0,
			})

		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.NoError(err)
		suite.Require().NotNil(token)
		suite.Equal("test", token.AccessToken)
	})

	suite.Run("error 500 unexpected content type", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(500).
			AddHeader("Content-Type", "text/html").
			BodyString("test")

		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrUnexpectedContentType)
		suite.Nil(token)
	})

	suite.Run("error 400 ESIA-007004", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(400).
			JSON(ErrorResponse{
				Error:            "access_denied",
				ErrorDescription: "ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос",
				State:            "test",
			})

		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrESIA_007004)
		suite.Nil(token)
	})

	suite.Run("error malformed json", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(200).
			AddHeader("Content-Type", "application/json").
			BodyString("not a json")

		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrUnmarshal)
		suite.Nil(token)
	})

	suite.Run("error guid", func() {
		guid = func() (string, error) {
			return "", ErrGUID
		}

		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrGUID)
		suite.Nil(token)
	})

	suite.Run("error request call", func() {
		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrRequestCall)
		suite.Nil(token)
	})

	suite.Run("error sign", func() {
		suite.client.signer = signature.NewNop("", "")
		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

	suite.Run("error signer is nil", func() {
		suite.client.signer = nil
		token, err := suite.client.TokenExchange("test", "test", "test")
		suite.ErrorIs(err, ErrTokenExchange)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

}

func (suite *suiteTestClient) TestTokenUpdate() {
	suite.Run("check request", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			MatchType("application/x-www-form-urlencoded").
			AddMatcher(matchFormField("client_id", "test")).
			AddMatcher(matchFormField("client_secret", base64.URLEncoding.EncodeToString([]byte(testSignature)))).
			AddMatcher(matchFormField("scope", "prm_chg?oid=test")).
			AddMatcher(matchFormField("client_certificate_hash", testCertHash)).
			AddMatcher(matchFormField("redirect_uri", "test")).
			AddMatcher(matchFormField("grant_type", "client_credentials")).
			AddMatcher(matchFormField("token_type", "Bearer")).
			AddMatcher(matchFormFieldRegexp("timestamp", `^\d{4}.\d{2}.\d{2} \d{2}:\d{2}:\d{2} [\+-]\d{4}$`)).
			AddMatcher(matchFormFieldRegexp("state", `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)). // guid
			Reply(200).
			JSON(TokenExchangeResponse{
				AccessToken: "test",
				IdToken:     "test",
				State:       "test",
				TokenType:   "Test",
				ExpiresIn:   0,
			})

		token, err := suite.client.TokenUpdate("test", "test")
		suite.NoError(err)
		suite.Require().NotNil(token)
		suite.Equal("test", token.AccessToken)
	})

	suite.Run("error 500 unexpected content type", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(500).
			AddHeader("Content-Type", "text/html").
			BodyString("test")

		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrUnexpectedContentType)
		suite.Nil(token)
	})

	suite.Run("error 400 ESIA-007004", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(400).
			JSON(ErrorResponse{
				Error:            "access_denied",
				ErrorDescription: "ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос",
				State:            "test",
			})

		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrESIA_007004)
		suite.Nil(token)
	})

	suite.Run("error malformed json", func() {
		gock.New("http://test.gock").
			Post(TokenEndpoint).
			Reply(200).
			AddHeader("Content-Type", "application/json").
			BodyString("not a json")

		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrUnmarshal)
		suite.Nil(token)
	})

	suite.Run("error guid", func() {
		guid = func() (string, error) {
			return "", ErrGUID
		}

		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrGUID)
		suite.Nil(token)
	})

	suite.Run("error request call", func() {
		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrRequestCall)
		suite.Nil(token)
	})

	suite.Run("error sign", func() {
		suite.client.signer = signature.NewNop("", "")
		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

	suite.Run("error signer is nil", func() {
		suite.client.signer = nil
		token, err := suite.client.TokenUpdate("test", "test")
		suite.ErrorIs(err, ErrTokenUpdate)
		suite.ErrorIs(err, ErrSign)
		suite.Nil(token)
	})

}

const (
	testSignature = "this is a test signature"
	testCertHash  = "test_hash"
)

func matchFormField(key, value string) gock.MatchFunc {
	return func(r *http.Request, _ *gock.Request) (bool, error) {
		if err := r.ParseForm(); err != nil {
			return false, err
		}
		got := r.Form.Get(key)
		if got == "" {
			return false, fmt.Errorf("form field %s not found", key)
		}
		if got != value {
			return false, fmt.Errorf("form field %s: expected %s, got %s", key, value, got)
		}
		return true, nil
	}
}

func matchFormFieldRegexp(key, regex string) gock.MatchFunc {
	return func(r *http.Request, _ *gock.Request) (bool, error) {
		if err := r.ParseForm(); err != nil {
			return false, err
		}
		got := r.Form.Get(key)
		if got == "" {
			return false, fmt.Errorf("form field %s not found", key)
		}

		re, err := regexp.Compile(regex)
		if err != nil {
			return false, err
		}
		if !re.MatchString(got) {
			return false, fmt.Errorf("form field %s: expected %s, got %s", key, re.String(), got)
		}

		return true, nil
	}
}

func notEmptyFormField(key string) gock.MatchFunc {
	return func(r *http.Request, _ *gock.Request) (bool, error) {
		if err := r.ParseForm(); err != nil {
			return false, err
		}
		if r.Form.Get(key) == "" {
			return false, fmt.Errorf("form field %s not found", key)
		}
		return true, nil
	}
}
