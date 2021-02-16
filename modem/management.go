package modem

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"
	"time"
)

type Logs struct {
	Operate bytes.Buffer
	Trace bytes.Buffer
}

func (c *Client) GetLogs() (*Logs, error) {
	token, err := c.getToken("/html/management/logcfg.asp")
	if err != nil { return nil, err }

	resp, err := c.httpPostForm("/html/management/logexport.log?RequestFile=success", token.form())
	if err != nil { return nil, err }
	files := tar.NewReader(bzip2.NewReader(resp.Body))

	logs := new(Logs)

	for {
		hdr, err := files.Next()
		if err != nil { break }
		switch hdr.Name {
		case "operateLog_export.txt":
			_, err = io.Copy(&logs.Operate, files)
			if err != nil { return nil, err }
		case "traceLog_export.txt":
			_, err = io.Copy(&logs.Trace, files)
			if err != nil { return nil, err }
		default:
			panic("unexpected log file: " + hdr.Name)
		}
	}
	return logs, nil
}

func (c *Client) DownloadConfigFile() (*bytes.Buffer, error) {
	token, err := c.getToken("/html/management/maintenance.asp")
	if err != nil { return nil, err }

	resp, err := c.httpPostForm("/html/management/downloadconfigfile.conf?RequestFile=success", token.form())
	if err != nil { return nil, err }

	return decrypt(resp.Body, cipherConf), nil
}

func (c *Client) Reboot() error {
	token, err := c.getToken("/html/management/maintenance.asp")
	if err != nil { return err }

	_, err = c.httpPostForm("/html/management/reboot.cgi", token.form())
	if err != nil { return err }

	time.Sleep(time.Minute * 1)

	for {
		resp, _ := c.httpPost("/index/getRebootRes.cgi", nil, nil)
		// there should not be any errors here
		if resp.Body.String() == "0" { break }
		fmt.Println("hello")
		time.Sleep(time.Second * 1)
	}

	return nil
}


