package aas

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
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
		return c.responseError(res)
	}

	//goland:noinspection ALL
	defer res.Body.Close()
	if !isJSONContentType(res.Header.Get("Content-Type")) {
		return fmt.Errorf("%w: '%s'", ErrUnexpectedContentType, res.Header.Get("Content-Type"))
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	return nil
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "application/json"
}
