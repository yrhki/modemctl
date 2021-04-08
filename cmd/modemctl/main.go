package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yrhki/modemctl/modem"
)

var client *modem.Client

func modemLogin(c *cli.Context) (err error) {
	username := c.String("username")
	password := c.String("password")
	url := c.String("url")
	client, err = modem.NewClient(url)
	if err != nil {
		return err
	}

	return client.Login(username, password)
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "username",
				Value:   "admin",
				Usage:   "Login username",
				EnvVars: []string{"MODEM_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "Login password",
				EnvVars: []string{"MODEM_PASSWORD"},
			},
			&cli.StringFlag{
				Name:    "url",
				Value:   "http://192.168.1.1",
				Usage:   "Modem URL",
				EnvVars: []string{"MODEM_URL"},
			},
		},
		Name: "modemctl",

		Commands: []*cli.Command{
			{
				Name:        "reboot",
				Category:    "management",
				Usage:       "reboot modem",
				Description: "reboot modem",

				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "wait",
					},
				},
			},

			{
				Name:        "ping",
				Category:    "management",
				Usage:       "ping ip",
				Description: "ping",
				Action:      ping,
			},
			{
				Name:        "traceroute",
				Category:    "management",
				Usage:       "traceroute ip",
				Description: "traceroute",
				Action:      traceroute,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
