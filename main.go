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
			&cli.DurationFlag{
				Name:    "default-timeout",
				Value:   2 * time.Second,
				EnvVars: []string{"DEFAULT_TIMEOUT"},
			},
		},
		Action: func(c *cli.Context) error {
			caps := selenium.Capabilities{}
			caps.AddChrome(chrome.Capabilities{})
			wd, err := selenium.NewRemote(caps, "http://localhost:9515")
			if err != nil {
				return fmt.Errorf("could not start web driver: %w", err)
			}

			wd.SetImplicitWaitTimeout(c.Duration("default-timeout"))

			defer wd.Close()

			rt := goja.New()
			rt.SetFieldNameMapper(goja.TagFieldNameMapper("goja", true))

			scriptName := c.String("script")
			scb, err := os.ReadFile(scriptName)
			if err != nil {
				return fmt.Errorf("could not read script %s: %w", scriptName, err)
			}

			err = rt.GlobalObject().Set("wd", &WDWrapper{wd: wd})
			if err != nil {
				return fmt.Errorf("could not set wd global: %w", err)
			}

			rt.GlobalObject().Set("println", fmt.Println)

			_, err = rt.RunScript(scriptName, string(scb))
			if err != nil {
				return fmt.Errorf("script failed: %w", err)
			}

			fmt.Println("script done")
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

func (w *WDWrapper) SetTimeout(timeout float64) {
	timeoutDuration := time.Duration(float64(time.Second) * timeout)
	w.wd.SetImplicitWaitTimeout(timeoutDuration)
}

func (w *WDWrapper) FindElement(selector string) (*ElementWrapper, error) {

	el, err := w.wd.FindElement(selenium.ByCSSSelector, selector)

	if err != nil {
		return nil, err
	}
	return &ElementWrapper{el}, nil
}

type ElementWrapper struct {
	el selenium.WebElement
}

func (e *ElementWrapper) Click() error {
	return e.el.Click()
}

func (e *ElementWrapper) Text() (string, error) {
	return e.el.Text()
}

func (e *ElementWrapper) SendKeys(text string) error {
	return e.el.SendKeys(text)
}
