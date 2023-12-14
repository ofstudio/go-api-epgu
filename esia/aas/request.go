package aas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) request(
	method,
	endpoint,
	contentType string,
	body io.Reader,
	result any,
) error {
	req, err := http.NewRequest(method, c.baseURI+endpoint, body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	c.logReq(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	c.logRes(res)

	if res.StatusCode >= 400 {
		return responseError(res)
	}

	//goland:noinspection ALL
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	return nil
}
