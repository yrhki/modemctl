package modem

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) httpPostForm(prefix string, body url.Values) (*bytes.Buffer, error) {
	req, err := http.NewRequest("POST", c.formatURL(prefix), strings.NewReader(body.Encode()))
	if err != nil { return nil, err }

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.c.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	fmt.Printf("%#+v\n", resp.Request.Header)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil { return nil, err }

	err = c.checkError(buf.String())
	if err != nil { return nil, err }

	return buf, nil
}

func (c *Client) httpPost(prefix string, headers http.Header, body io.Reader) (*bytes.Buffer, error) {
	req, err := http.NewRequest("POST", c.formatURL(prefix), body)
	if err != nil { return nil, err }

	if headers != nil { req.Header = headers }

	resp, err := c.c.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	fmt.Printf("%#+v\n", resp)


	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil { return nil, err }

	err = c.checkError(buf.String())
	if err != nil { return nil, err }

	return buf, nil
}

func (c *Client) httpGet(prefix string) (*bytes.Buffer, error) {
	resp, err := c.c.Get(c.formatURL(prefix))
	if err != nil { return nil, err }
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil { return nil, err }

	err = c.checkError(buf.String())
	if err != nil { return nil, err }

	return buf, nil
}


