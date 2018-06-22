package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli"
)

// UserHomeDir is a function that returns the user's home directory.
func UserHomeDir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env)
}

func main() {
	var trigger string
	var computer string
	var params string
	var urlparams string

	dir := UserHomeDir()
	// fmt.Println(dir)

	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Name = "tcmd"
	app.Usage = "Run commands on computers in your TRIGGERcmd account"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "trigger, t",
			Usage:       "Trigger name of the command you want to run",
			Destination: &trigger,
		},
		cli.StringFlag{
			Name:        "computer, c",
			Usage:       "Name of the computer (leave blank for your default computer)",
			Destination: &computer,
		},
		cli.StringFlag{
			Name:        "params, p",
			Usage:       "Any parameters you want to add to the remote command",
			Destination: &params,
		},
	}

	app.Action = func(c *cli.Context) error {
		if trigger == "" {
			fmt.Println("No trigger specified.  Use --help or -h for help.")
		} else {
			// fmt.Println(strings.Join("Trigger: ", trigger))
			t := []string{urlparams, "&trigger=", trigger}
			urlparams = strings.Join(t, "")

			if computer == "" {
				// fmt.Println("No computer specified.  Using default computer.")
			} else {
				// fmt.Println(strings.Join("Computer: ", computer))
				s := []string{urlparams, "&computer=", computer}
				urlparams = strings.Join(s, "")
			}

			if params == "" {
				// fmt.Println("No parameters specified.")
			} else {
				// fmt.Println(strings.Join("Parameters: ", params))
				s := []string{urlparams, "&params=", params}
				urlparams = strings.Join(s, "")
			}

			p := filepath.Join(dir, "/.TRIGGERcmdData/token.tkn")
			// fmt.Println(p)

			b, err := ioutil.ReadFile(p) // just pass the file name
			if err != nil {
				fmt.Print(err)
			}

			token := string(b) // convert content to a 'string'

			s := []string{"https://www.triggercmd.com/api/run/triggersave?token=", token, urlparams}
			// fmt.Println(strings.Join(s, ""))

			resp, err := http.Get(strings.Join(s, ""))
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", body)
		}
		return nil
	}

	cli.AppHelpTemplate = `NAME:
    {{.Name}} - {{.Usage}}
  USAGE:
    {{.HelpName}} {{if .VisibleFlags}}[options]{{end}}
    {{if len .Authors}}
  AUTHOR:
    {{range .Authors}}{{ . }}{{end}}
    {{end}}{{if .Commands}}
  OPTIONS:
    {{range .VisibleFlags}}{{.}}
    {{end}}{{end}}
  `

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
