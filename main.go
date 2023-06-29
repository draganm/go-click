package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dop251/goja"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "script",
				Value:   "script.js",
				EnvVars: []string{"SCRIPT"},
			},
		},
		Action: func(c *cli.Context) error {
			caps := selenium.Capabilities{}
			caps.AddChrome(chrome.Capabilities{})
			wd, err := selenium.NewRemote(caps, "http://localhost:9515")
			if err != nil {
				return fmt.Errorf("could not start web driver: %w", err)
			}

			rt := goja.New()
			rt.SetFieldNameMapper(goja.TagFieldNameMapper("goja", true))

			scriptName := c.String("script")
			scb, err := os.ReadFile(scriptName)
			if err != nil {
				return fmt.Errorf("could not read script %s: %w", scriptName, err)
			}

			err = rt.GlobalObject().Set("wd", &WDWrapper{wd})
			if err != nil {
				return fmt.Errorf("could not set wd global: %w", err)
			}
			_, err = rt.RunScript(scriptName, string(scb))
			if err != nil {
				return fmt.Errorf("script failed: %w", err)
			}

			fmt.Println("script done")
			// err = wd.Get("http://www.google.com")
			// if err != nil {
			// 	return err
			// }
			// // wd.FindElement()
			time.Sleep(20 * time.Second)
			return nil
		},
	}
	app.RunAndExitOnError()

}

type WDWrapper struct {
	wd selenium.WebDriver
}

func (w *WDWrapper) Get(url string) error {
	return w.wd.Get(url)
}
