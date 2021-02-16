package modem

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)


type token struct {
	csrfToken, csrfParam string
}

func (t *token) form() url.Values {
	return url.Values{
		"csrf_param": {t.csrfParam},
		"csrf_token": {t.csrfToken},
	}
}

type Client struct {
	c http.Client
	url *url.URL
	username, password string
	sessionID string
}

func NewClient(modemURL string) (*Client, error) {
	var err error
	c := new(Client)
	c.url, err = url.Parse(modemURL)
	if err != nil { return nil, err }

	jar, err := cookiejar.New(nil)
	if err != nil { return nil, err }
	jar.SetCookies(c.url, []*http.Cookie{
			{
				Name: "Language",
				Value: "en_us",
			},
		},
	)

	c.c = http.Client{
		Jar:jar,
	}
	return c, nil
}

func (c *Client) formatURL(path string) string {
	if path == "" {
		return fmt.Sprintf("%s://%s", c.url.Scheme, c.url.Host)
	}
	return fmt.Sprintf("%s://%s%s", c.url.Scheme, c.url.Host, path)
}

func (c *Client) getToken(prefix string) (*token, error) {
	resp, err := c.httpGet(prefix)
	if err != nil { return nil, err }

	t := new(token)

	csrfParamReg := regexp.MustCompile(`var csrf_param = "(.*)"`)
	csrfTokenReg := regexp.MustCompile(`var csrf_token = "(.*)"`)
	matches := csrfParamReg.FindStringSubmatch(resp.Body.String())
	t.csrfParam = matches[1]
	matches = csrfTokenReg.FindStringSubmatch(resp.Body.String())
	t.csrfToken = matches[1]

	return t, nil
}

func (c *Client) getRedirect(page string) (string, error) {
	pageNameReg := regexp.MustCompile(`var pageName = '(.*)';`)
	matches := pageNameReg.FindStringSubmatch(page)
	if len(matches) < 2 {
		return "", nil
	}
	return matches[1], nil
}

func (c *Client) GetEventStatus() error {
	resp, err := c.httpGet("/html/ajaxref/getEventStatus.cgi")
	if err != nil { return err }
	fmt.Println(resp.Body.String())
	return nil
}

func (c *Client) Info() error {
	resp, err := c.c.Get(c.formatURL("html/status/systeminfo.asp"))
	if err != nil { return err }
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil { return err }

	r := regexp.MustCompile(`systemInfoObj = {'sn' : '(.*)', 'prodctname' : '(.*)', 'swbuildtime' : '(.*)', 'swver' : '(.*)', 'hwver' : '(.*)'};\nIMEI = '(.*)';`)
	fmt.Println(r.FindStringSubmatch(string(b))[1:])

	return nil
}
