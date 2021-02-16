package modem

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type httpResponse struct {
	Body *bytes.Buffer
}

func (c *Client) httpRequest(req *http.Request) (*httpResponse, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:86.0) Gecko/20100101 Firefox/86.0")
	req.Header.Set("Host", "192.168.1.1")

	resp, err := c.c.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	httpResp := &httpResponse{
		Body: new(bytes.Buffer),
	}

	_, err = io.Copy(httpResp.Body, resp.Body)
	if err != nil { return httpResp, err }

	err = c.checkError(httpResp.Body)
	if err != nil { return httpResp, err }

	return httpResp, nil
}

func (c *Client) httpPostForm(prefix string, body url.Values) (*httpResponse, error) {
	req, err := http.NewRequest("POST", c.formatURL(prefix), strings.NewReader(body.Encode()))
	if err != nil { return nil, err }

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.httpRequest(req)
}

func (c *Client) httpPost(prefix string, headers http.Header, body io.Reader) (*httpResponse, error) {
	req, err := http.NewRequest("POST", c.formatURL(prefix), body)
	if err != nil { return nil, err }

	if headers != nil { req.Header = headers }

	return c.httpRequest(req)
}

func (c *Client) httpGet(prefix string) (*httpResponse, error) {
	req, err := http.NewRequest("GET", c.formatURL(prefix), nil)
	if err != nil { return nil, err }

	return c.httpRequest(req)
}


