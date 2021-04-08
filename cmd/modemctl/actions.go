package main

import (
	"github.com/urfave/cli/v2"
)

func ping(c *cli.Context) error {
	target := "1.1.1.1"
	if c.NArg() > 0 {
		target = c.Args().Get(0)
	}
	err := modemLogin(c)
	if err != nil {
		return err
	}
	err = client.DiagnosisPing(target, 32, 4, false)
	if err != nil {
		return err
	}

	ok, err := client.DiagnosisPingResult(c.App.Writer)
	if err != nil {
		return err
	}

	if !ok {
		return cli.Exit("", 1)
	}

	return nil
}

func traceroute(c *cli.Context) error {
	target := "1.1.1.1"
	if c.NArg() > 0 {
		target = c.Args().Get(0)
	}
	err := modemLogin(c)
	if err != nil {
		return err
	}
	err = client.DiagnosisTraceroute(target, 30, 4)
	if err != nil {
		return err
	}

	ok, err := client.DiagnosisPingResult(c.App.Writer)
	if err != nil {
		return err
	}

	if !ok {
		return cli.Exit("", 1)
	}

	return nil
}
