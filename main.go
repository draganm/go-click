package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dop251/goja"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
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

			eg, ctx := errgroup.WithContext(context.Background())

			eg.Go(func() error {

				defer func() {
					x := recover()
					if x != nil {
						fmt.Println(x)
					}
					fmt.Println("defer done")
				}()

				caps := selenium.Capabilities{}
				caps.AddChrome(chrome.Capabilities{})
				wd, err := selenium.NewRemote(ctx, caps, "http://localhost:9515")
				if err != nil {
					return fmt.Errorf("could not start web driver: %w", err)
				}

				wd.SetImplicitWaitTimeout(c.Duration("default-timeout"))

				defer wd.Quit()

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

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(30 * time.Second):
					return nil
				}
			})

			eg.Go(func() error {

				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case sig := <-sigs:
					fmt.Println("signal received, terminating", "sig", sig)
					return fmt.Errorf("signal %s received", sig.String())
				}

			})

			defer func() {
				time.Sleep(200 * time.Millisecond)
			}()
			return eg.Wait()
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
