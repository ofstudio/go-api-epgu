package apipgu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) requestJSON(
	ctx context.Context,
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
	result any,
) error {
	resBody, err := c.requestBody(ctx, method, endpoint, contentType, accessToken, body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(resBody, result); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	return nil
}

func (c *Client) requestBody(
	ctx context.Context,
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
) ([]byte, error) {
	resBody := &bytes.Buffer{}
	if err := c.requestStream(ctx, method, endpoint, contentType, accessToken, body, resBody); err != nil {
		return nil, err
	}
	return resBody.Bytes(), nil
}

func (c *Client) requestStream(
	ctx context.Context,
	method,
	endpoint,
	contentType,
	accessToken string,
	body io.Reader,
	dst io.Writer,
) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURI+endpoint, body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
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
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}
	defer res.Body.Close()

	c.logRes(res)

	if res.StatusCode >= 400 || res.StatusCode == http.StatusNoContent {
		return responseError(res)
	}

	_, err = io.Copy(dst, res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	return nil
}
