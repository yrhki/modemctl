package modem

import (
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WirelessStatus struct {
	PLMN int
	Status string
	// dBm
	RSSI float32
	// dBm
	RSRP float32
	// dB
	RSRQ float32
	Roaming bool
	Band string
}

func (c *Client) DiagnosisWirelessStatus() (*WirelessStatus, error) {
	resp, err := c.httpPost(fmt.Sprintf("/index/getmodemsts.cgi?rid=%v", rand.Float64()), nil, nil)
	if err != nil { return nil, err }

	values := strings.Split(resp.String(), ",")
	fmt.Printf("%#+v\n", values)

	plmn, err := strconv.Atoi(values[1])
	if err != nil { return nil, err }

	return &WirelessStatus{
		PLMN: plmn,
		Band: values[7],
	}, nil
}

func (c *Client) DiagnosisPing(target string, packetSize, timeout int, fragment bool) error {
	u, err := url.Parse(c.formatURL("/html/management/ping.cgi"))
	if err != nil { return err }

	q := u.Query()
	q.Set("target", target)
	q.Set("packetsize", fmt.Sprint(packetSize))
	q.Set("timeout", fmt.Sprint(timeout))
	q.Set("RequestFile", "success")
	if fragment { q.Set("fragment", "1")
	} else { q.Set("fragment", "0") }
	u.RawQuery = q.Encode()

	token, err := c.getToken("/html/management/diagnose.asp")
	if err != nil { return err }


	_, err = c.httpPostForm(fmt.Sprintf("%s?%s", u.Path, u.RawQuery), token.form())
	if err != nil { return err }

	time.Sleep(time.Second * time.Duration(timeout))
	return nil
}

func (c *Client) DiagnosisTraceroute(target string, maxHops, timeout int) error {
	u, err := url.Parse(c.formatURL("/html/management/traceroute.cgi"))
	if err != nil { return err }

	q := u.Query()
	q.Set("target", target)
	q.Set("maxhops", fmt.Sprint(maxHops))
	q.Set("timeout", fmt.Sprint(timeout))
	q.Set("RequestFile", "success")
	u.RawQuery = q.Encode()

	token, err := c.getToken("/html/management/diagnose.asp")
	if err != nil { return err }


	_, err = c.httpPostForm(fmt.Sprintf("%s?%s", u.Path, u.RawQuery), token.form())
	if err != nil { return err }

	time.Sleep(time.Second * time.Duration(timeout))
	return nil
}

// Result of DiagnosisPing or DiagnosisTraceroute
func (c *Client) DiagnosisPingResult(output io.Writer) (bool, error) {
	resp, err := c.httpGet("/html/management/pingresult.asp")
	if err != nil { return false, err }

	lines := strings.Split(resp.String(), " + ")
	lines = lines[:len(lines)-1]

	if !strings.HasPrefix(lines[len(lines)-1], "\"__") {
		time.Sleep(time.Second * 1)
		return c.DiagnosisPingResult(output)
	}

	for _, line := range lines[:len(lines)-1] {
		line = line[1:len(line)-1]
		line = strings.Replace(line, "\\n", "\n", 1)
		fmt.Fprint(output, line)
	}

	return lines[len(lines)-1] == `"__finshed__\n"`, nil
}


