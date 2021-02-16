package modem

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	regLoginError      = regexp.MustCompile(`var LoginError = "(.*)";`)
	regLoginCookieFlag = regexp.MustCompile(`var Cookieflag = (\d*);`)
	regRedirect        = regexp.MustCompile(`var pageName = '(.*)';`)

	loginErrorMessages = []string{
		"Login failed. Enable Cookies on your browser.",
		"Login failed. Another user has already logged in using this account. Please try again later.",
		"Enter your username and password",
		"Login failed. You can try two more times.",
		"Login failed. You can try one more times.",
		"You have attempted to log in three consecutive times unsuccessfully. Please wait one minute before retrying.",
	}
)

func (c *Client) checkError(body *bytes.Buffer) error {
	matches := regRedirect.FindStringSubmatch(body.String())
	if len(matches) > 0 {
		switch matches[1] {
		case "/html/msgerrcode.asp":
			return errors.New("invalid parameters")
		case "success", "/html/status/overview.asp":
			// No error
		case "/":
			resp, err := c.httpGet("/")
			if err != nil { return err }
			matches = regLoginError.FindStringSubmatch(resp.Body.String())

			if len(matches) != 2 { return errors.New("no login error") }

			loginErr := strings.Split(matches[1], ":")

			loginTimes, err := strconv.Atoi(loginErr[0])
			if err != nil { return err }
			// loginErrorCode, err := strconv.Atoi(loginErr[1])
			// if err != nil { return err }

			matches = regLoginCookieFlag.FindStringSubmatch(resp.Body.String())
			cookieFlag, err := strconv.Atoi(matches[1])
			if err != nil { return err }

			switch cookieFlag {
			case 1:
				return errors.New(loginErrorMessages[0])
			case 2:
				return errors.New(loginErrorMessages[1])
			}

			switch loginTimes {
			case 1:
				return errors.New(loginErrorMessages[2])
			case 2:
				return errors.New(loginErrorMessages[3])
			case 3:
				return errors.New(loginErrorMessages[4])
			default:
				return errors.New(loginErrorMessages[5])
			}
		default:
			panic("Unhandled redirect: "+matches[1])
		}
	}
	return nil
}


