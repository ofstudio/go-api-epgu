package apipgu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) requestJSON(
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
	result any,
) error {
	resBody, err := c.requestBody(method, endpoint, contentType, accessToken, body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	return nil
}

func (c *Client) requestBody(
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURI+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	c.logReq(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	c.logRes(res)

	if res.StatusCode >= 400 || res.StatusCode == http.StatusNoContent {
		return nil, responseError(res)
	}

	//goland:noinspection ALL
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	return resBody, nil
}
