package modem

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tebeka/selenium"
)

func waitDocumentReady(wd selenium.WebDriver) (bool, error) {
	state, err := wd.ExecuteScript(`return document.readyState;`, nil)
	if err != nil { return false, err }

	if state.(string) == "complete" {
		return true, nil
	}
	return false, nil
}




func (c *Client) Login(username, password string) error {
	const port = 8080
	seleniumOpts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
	}

	service, err := selenium.NewSeleniumService("/usr/share/selenium-server/selenium-server-standalone.jar", port, seleniumOpts...)
	if err != nil { return err }
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil { return err }
	defer wd.Close()
	defer wd.Quit()

	err = wd.Get(c.url.String())
	if err != nil { panic(err) }

	wd.Wait(waitDocumentReady)

	//elemUsername, err := wd.FindElement(selenium.ByCSSSelector, "#txt_Username")
	//if err != nil { panic(err) }
	//err = elemUsername.SendKeys(username)
	//if err != nil { panic(err) }

	elemPassword, err := wd.FindElement(selenium.ByCSSSelector, "#txt_Password")
	if err != nil { panic(err) }
	err = elemPassword.SendKeys(password)
	if err != nil { panic(err) }


	_, err = wd.ExecuteScript(`document.getElementById("login_btn").click();`, nil)
	if err != nil { panic(err) }

	//userAgent, err := wd.ExecuteScript(`return navigator.userAgent;`, nil)
	//if err != nil { panic(err) }

	if e, _ := wd.FindElement(selenium.ByCSSSelector, "#erroinfoId"); e != nil {
		errText, err := e.Text()
		if err != nil { panic(err) }
		return errors.New(errText)
	}
	cookie, err := wd.GetCookie("SessionID_R3")
	if err != nil { panic(err) }
	c.sessionID = cookie.Value

	c.c.Jar.SetCookies(c.url, []*http.Cookie{
		{
			Name: cookie.Name,
			Value: cookie.Value,
		},
	})

	return nil
}

