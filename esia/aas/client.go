package aas

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ofstudio/go-api-epgu/esia/signature"
	"github.com/ofstudio/go-api-epgu/utils"
)

const tsLayout = "2006.01.02 15:04:05 -0700"

const (
	UserEndpoint  = "/aas/oauth2/v2/ac" // URI страницы ЕСИА для предоставления пользователем запрошенных прав
	TokenEndpoint = "/aas/oauth2/v3/te" // Эндпоинт для обмена кода авторизации на маркер доступа
)

// ErrorResponse - ответ от ЕСИА при ошибке
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	State            string `json:"state"`
}

// TokenExchangeResponse - ответ от ЕСИА при успешном обмене кода на маркер доступа
type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	State       string `json:"state"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Client - OAuth2-клиент для запроса согласия и маркера доступа ЕСИА
// для получателей услуг ЕПГУ - физических лиц.
type Client struct {
	baseURI    string
	clientId   string
	signer     signature.Provider
	httpClient *http.Client
	logger     utils.Logger
	debug      bool
}

// NewClient - конструктор для Client.
// Параметры:
//   - baseURI - URI ЕСИА
//   - clientId - мнемоника ИС-потребителя
//   - signer - провайдер подписи запросов
func NewClient(baseURI, clientId string, signer signature.Provider) *Client {
	return &Client{
		baseURI:    baseURI,
		clientId:   clientId,
		signer:     signer,
		httpClient: &http.Client{},
	}
}

// WithDebug - включает логирование запросов и ответов
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

// WithHTTPClient - устанавливает http-клиент для запросов к ЕСИА
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	if httpClient != nil {
		c.httpClient = httpClient
	}
	return c
}

// AuthURI - формирует URI на страницу ЕСИА для предоставления пользователем запрошенных прав.
// Тк используется параметр [Permissions], то в scope необходимо указывать "openid".
//
// Возвращает URI на страницу ЕСИА либо цепочку ошибок из [ErrAuthURI] и других:
//   - [ErrSign] - ошибка подписи ссылки
//   - [ErrGUID] - при невозможности сформировать GUID
//
// Подробнее см "Методические рекомендации по использованию ЕСИА",
// раздел "Получение авторизационного кода (v2/ac)".
func (c *Client) AuthURI(scope, redirectURI string, permissions Permissions) (string, error) {
	timestamp := time.Now().UTC().Format(tsLayout)
	state, err := guid()
	if err != nil {
		return "", fmt.Errorf("%w: %w: %w", ErrAuthURI, ErrGUID, err)
	}
	clientSecret, err := c.sign(c.clientId, scope, timestamp, state, redirectURI)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAuthURI, err)
	}

	params := &url.Values{}
	params.Add("client_id", c.clientId)
	params.Add("client_secret", clientSecret)
	params.Add("scope", scope)
	params.Add("timestamp", timestamp)
	params.Add("state", state)
	params.Add("redirect_uri", redirectURI)
	params.Add("client_certificate_hash", c.signer.CertHash())
	params.Add("response_type", "code")
	params.Add("access_type", "online")
	params.Add("permissions", permissions.Base64String())

	return c.baseURI + UserEndpoint + "?" + params.Encode(), nil
}

// ParseCallback - возвращает код авторизации code и state из
// query-параметров callback-запроса к redirect_uri от ЕСИА.
//
// Подробнее см "Методические рекомендации по использованию ЕСИА",
// раздел "Получение авторизационного кода (v2/ac)".
//
// В случае ошибки возвращает цепочку из [ErrParseCallback] и других:
//   - [ErrNoState] - отсутствует параметр state
//   - [ErrParseCallback] - ошибка разбора параметров
//   - ошибка ЕСИА: ErrESIAxxxxxx ([ErrESIA_007003] и др.)
//
// Пример сообщения об ошибке:
//
//	ESIA-007014: Запрос не содержит обязательного параметра [error='invalid_request', error_description='ESIA-007014: The request does not contain the mandatory parameter' state='48d1a8dc-0b7d-418a-b4ef-2c7797f77dc9']'
func (c *Client) ParseCallback(query url.Values) (string, string, error) {
	state := query.Get("state")
	if state == "" {
		return "", "", fmt.Errorf("%w: %w", ErrParseCallback, ErrNoState)
	}

	code := query.Get("code")
	if code == "" {
		return "", state, fmt.Errorf("%w: %w", ErrParseCallback, callbackError(query))
	}

	return code, state, nil
}

// TokenExchange обменивает код авторизации на маркер доступа.
// Параметры scope и redirectURI должны быть такими же, как и при вызове [Client.AuthURI].
//
// Подробнее см "Методические рекомендации по использованию ЕСИА",
// раздел "Получение маркера доступа в обмен на авторизационный код (v3/te)".
//
// Возвращает ответ от ЕСИА [TokenExchangeResponse] либо цепочку ошибок из [ErrTokenExchange] и других:
//   - [ErrSign] - ошибка подписи запроса
//   - [ErrGUID] - при невозможности сформировать GUID
//   - [ErrRequest] - ошибка HTTP-запроса
//   - [ErrJSONUnmarshal] - ошибка разбора ответа
//   - [ErrUnexpectedContentType] - неожидаемый Content-Type ответа
//   - ошибок ЕСИА ErrESIA_xxxxxx ([ErrESIA_007004] и др.)
//
// Пример сообщения об ошибке:
//
//	HTTP 400 Bad request: ESIA-007014: Запрос не содержит обязательного параметра [error='invalid_request', error_description='ESIA-007014: The request does not contain the mandatory parameter' state='48d1a8dc-0b7d-418a-b4ef-2c7797f77dc9']'
func (c *Client) TokenExchange(code, scope, redirectURI string) (*TokenExchangeResponse, error) {
	timestamp := time.Now().UTC().Format(tsLayout)
	state, err := guid()
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrTokenExchange, ErrGUID, err)
	}
	clientSecret, err := c.sign(c.clientId, scope, timestamp, state, redirectURI, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	reqBody := url.Values{}
	reqBody.Set("client_id", c.clientId)
	reqBody.Set("client_secret", clientSecret)
	reqBody.Set("scope", scope)
	reqBody.Set("timestamp", timestamp)
	reqBody.Set("state", state)
	reqBody.Set("redirect_uri", redirectURI)
	reqBody.Set("client_certificate_hash", c.signer.CertHash())
	reqBody.Set("code", code)
	reqBody.Set("grant_type", "authorization_code")
	reqBody.Set("token_type", "Bearer")

	result := &TokenExchangeResponse{}

	if err = c.request(
		http.MethodPost,
		TokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(reqBody.Encode()),
		result,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchange, err)
	}

	return result, nil
}

// TokenUpdate обновляет маркер доступа по идентификатору пользователя (OID),
// используя scope="prm_chg". Параметр redirectURI должен быть таким же, как и при вызове AuthURI.
//
// Подробнее см "Методические рекомендации по интеграции с REST API Цифрового профиля"
// раздел "Online-режим запроса согласий".
//
// Возвращает ответ от ЕСИА [TokenExchangeResponse] либо цепочку ошибок из [ErrTokenUpdate] и
// ошибок аналогичных TokenExchange.
func (c *Client) TokenUpdate(oid, redirectURI string) (*TokenExchangeResponse, error) {
	timestamp := time.Now().UTC().Format(tsLayout)
	scope := "prm_chg?oid=" + oid
	state, err := guid()
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrTokenUpdate, ErrGUID, err)
	}
	clientSecret, err := c.sign(c.clientId, scope, timestamp, state, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenUpdate, err)
	}

	reqBody := url.Values{}
	reqBody.Set("client_id", c.clientId)
	reqBody.Set("client_secret", clientSecret)
	reqBody.Set("scope", scope)
	reqBody.Set("timestamp", timestamp)
	reqBody.Set("state", state)
	reqBody.Set("redirect_uri", redirectURI)
	reqBody.Set("client_certificate_hash", c.signer.CertHash())
	reqBody.Set("grant_type", "client_credentials")
	reqBody.Set("token_type", "Bearer")

	result := &TokenExchangeResponse{}
	if err = c.request(
		http.MethodPost,
		TokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(reqBody.Encode()),
		result,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenUpdate, err)
	}
	return result, nil
}

func (c *Client) sign(args ...string) (string, error) {
	if c.signer == nil {
		return "", fmt.Errorf("%w: signer not specified", ErrSign)
	}
	data := []byte(strings.Join(args, ""))
	sign, err := c.signer.Sign(data)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSign, err)
	}
	return base64.URLEncoding.EncodeToString(sign), nil
}

func (c *Client) logReq(req *http.Request) {
	if c.debug {
		utils.LogReq(req, c.logger)
	}
}

func (c *Client) logRes(res *http.Response) {
	if c.debug {
		utils.LogRes(res, c.logger)
	}
}

var guid = utils.GUID
