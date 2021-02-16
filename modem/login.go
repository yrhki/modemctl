package modem

import (
	"net/url"
)

func (c *Client) Login(username, password string) error {
	t, err := c.getToken("")
	if err != nil { return err }

	pass, err := encryptPassword(username, password, t)
	if err != nil { return err }
	v := url.Values{
		"Username":        {username},
		"Password":        {pass},
		"csrf_param":        {t.csrfParam},
		"csrf_token":        {t.csrfToken},
	}

	_, err = c.httpPostForm("/index/login.cgi", v)
	if err != nil { return err }
	c.loggedIn = true
	return nil
}

func (c *Client) Logout() error {
	t, err := c.getToken("")
	if err != nil { return err }

	_, err = c.httpPostForm("/index/logout.cgi", t.form())
	if err != nil && err != ErrNoLogin { return err }
	c.loggedIn = false
	return nil
}
