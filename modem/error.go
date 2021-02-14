package modem

import (
	"errors"
	"regexp"
)


func (c *Client) checkError(text string) error {
	pageNameReg := regexp.MustCompile(`var pageName = '(.*)';`)
	matches := pageNameReg.FindStringSubmatch(text)
	if len(matches) > 0 {
		switch matches[1] {
		case "/html/msgerrcode.asp":
			return errors.New("invalid parameters")
		case "success":
			// No error
		case "/":
			return errors.New("Redirected to login page")
		default:
			panic("Unhandled redirect: "+matches[1])
		}
	}
	return nil
}


