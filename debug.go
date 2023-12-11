package apipgu

import (
	"net/http"

	"github.com/ofstudio/go-api-epgu/utils"
)

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
