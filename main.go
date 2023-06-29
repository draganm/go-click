package main

import (
	"github.com/tebeka/selenium"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Action: func(c *cli.Context) error {
			selenium.NewRemote()
			return nil
		},
	}
	app.RunAndExitOnError()

}
