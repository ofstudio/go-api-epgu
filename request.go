package apipgu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) request(
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
	result any,
) error {
	req, err := http.NewRequest(method, c.baseURI+endpoint, body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestPrepare, err)
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
		return fmt.Errorf("%w: %w", ErrRequestCall, err)
	}

	c.logRes(res)

	if res.StatusCode >= 400 || res.StatusCode == http.StatusNoContent {
		return responseError(res)
	}

	//goland:noinspection ALL
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResponseRead, err)
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	return nil
}
