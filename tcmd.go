package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	var list bool
	var pair bool
	var trigger string
	var computer string
	var params string
	var urlparams string
	var panel string
	var button string
	var listpanels bool

	var pairResult map[string]interface{}
	var pairLookupResult map[string]interface{}
	var pairToken string

	dir := UserHomeDir()

	app := cli.NewApp()
	app.Version = "1.0.6"
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
		cli.StringFlag{
			Name:        "panel, P",
			Usage:       "Name of the panel you want to use",
			Destination: &panel,
		},
		cli.StringFlag{
			Name:        "button, b",
			Usage:       "Name of the panel button to \"press\"",
			Destination: &button,
		},
		cli.BoolFlag{
			Name:        "list, l",
			Usage:       "List your commands",
			Destination: &list,
		},
		cli.BoolFlag{
			Name:        "listpanels, L",
			Usage:       "List your panels",
			Destination: &listpanels,
		},
		cli.BoolFlag{
			Name:        "pair",
			Usage:       "Login using a pair code",
			Destination: &pair,
		},
	}

	app.Action = func(c *cli.Context) error {
		p := filepath.Join(dir, "/.TRIGGERcmdData/token.tkn")

		b, err := ioutil.ReadFile(p) // just pass the file name
		if err != nil {
			fmt.Print(err)
		}

		token := string(b) // convert content to a 'string'

		if pair {
			if token == "" {
				fmt.Println("\nNo token found.\nWithin 10 minutes, log into your account at triggercmd.com, click your name in the upper-right, click Pair, and type in this pair code:")

				s := []string{"https://www.triggercmd.com/pair"}

				resp, err := http.Get(strings.Join(s, ""))
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					panic(err)
				}

				json.Unmarshal([]byte(body), &pairResult)
				fmt.Printf("%s\n", pairResult["pairCode"])
				fmt.Printf("\n%s", "Waiting..")
				pairToken = fmt.Sprintf("%v", pairResult["pairToken"])

				timeout := time.After(10 * time.Minute)
				ticker := time.Tick(5000 * time.Millisecond)

			forLoop:
				for {
					select {
					case <-timeout:
						fmt.Printf("%s", "\nTimed out.  Run tcmd --pair again to get a new pair code.")
						break forLoop
					case <-ticker:
						s = []string{"https://www.triggercmd.com/pair/lookup?token=", pairToken}
						resp, err = http.Get(strings.Join(s, ""))
						if err != nil {
							panic(err)
						}
						defer resp.Body.Close()
						body, err = ioutil.ReadAll(resp.Body)
						if err != nil {
							panic(err)
						}
						// fmt.Printf("%s", body)
						json.Unmarshal([]byte(body), &pairLookupResult)
						_, ok := pairLookupResult["token"]
						if ok {
							token = fmt.Sprintf("%v", pairLookupResult["token"])
							err := os.MkdirAll(filepath.Join(dir, "/.TRIGGERcmdData"), os.ModeDir)
							if err == nil {
								tokendata := []byte(token)
								err := ioutil.WriteFile(p, tokendata, 0700)
								if err == nil {
									fmt.Printf("%s", "Token saved.\nGo ahead and run something like:  tcmd --list")
								} else {
									fmt.Printf("%s", "Something went wrong while creating the ~/.TRIGGERcmdData/token.tkn file.")
								}
							} else {
								fmt.Printf("%s", "Something went wrong while creating the .TRIGGERcmdData directory in your home directory.")
							}

							break forLoop
						} else {
							fmt.Printf("%s", ".")
						}
					}
				}

			} else {
				fmt.Println("You already have a token.  There's no need to pair.")
			}
		} else {
			if token == "" {
				fmt.Println("\nNo token found.  Install the TRIGGERcmd agent, or use --pair to get a token.")
			} else {
				if list {
					s := []string{"https://www.triggercmd.com/api/command/list?token=", token}

					resp, err := http.Get(strings.Join(s, ""))
					if err != nil {
						panic(err)
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						panic(err)
					}

					var outputline string
					var result map[string]interface{}
					json.Unmarshal([]byte(body), &result)
					commands := result["records"].([]interface{})

					for key, value := range commands {
						if value.(map[string]interface{})["allowParams"] == true {
							t := []string{
								"tcmd --computer \"",
								value.(map[string]interface{})["computer"].(map[string]interface{})["name"].(string),
								"\" --trigger \"",
								value.(map[string]interface{})["name"].(string),
								"\" --params \"(your parameters)\""}
							outputline = strings.Join(t, "")
							fmt.Println(key, outputline)
						} else {
							t := []string{
								"tcmd --computer \"",
								value.(map[string]interface{})["computer"].(map[string]interface{})["name"].(string),
								"\" --trigger \"",
								value.(map[string]interface{})["name"].(string),
								"\""}
							outputline = strings.Join(t, "")
							fmt.Println(key, outputline)
						}
					}
				} else {
					if listpanels {
						s := []string{"https://triggercmd.com/api/panelbutton/list?token=", token}

						resp, err := http.Get(strings.Join(s, ""))
						if err != nil {
							panic(err)
						}
						defer resp.Body.Close()
						body, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							panic(err)
						}

						var outputline string
						var result map[string]interface{}
						json.Unmarshal([]byte(body), &result)
						buttons := result["records"].([]interface{})

						for key, value := range buttons {
							t := []string{
								"tcmd --panel \"",
								value.(map[string]interface{})["panel"].(map[string]interface{})["name"].(string),
								"\" --button \"",
								value.(map[string]interface{})["name"].(string),
								"\""}
							outputline = strings.Join(t, "")
							fmt.Println(key, outputline)
						}
					} else {
						if panel != "" {
							if button == "" {
								fmt.Println("No button specified.  Use --help or -h for help.")
							} else {
								t := []string{urlparams, "&button=", url.PathEscape(button)}
								urlparams = strings.Join(t, "")

								p := []string{urlparams, "&panel=", url.PathEscape(panel)}
								urlparams = strings.Join(p, "")

								if params == "" {
									// fmt.Println("No parameters specified.")
								} else {
									s := []string{urlparams, "&params=", url.PathEscape(params)}
									urlparams = strings.Join(s, "")
								}

								s := []string{"https://www.triggercmd.com/api/panel/trigger?token=", token, urlparams}
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
						} else {
							if trigger == "" {
								fmt.Println("No trigger specified.  Use --help or -h for help.")
							} else {
								t := []string{urlparams, "&trigger=", url.PathEscape(trigger)}
								urlparams = strings.Join(t, "")

								if computer == "" {
									// fmt.Println("No computer specified.  Using default computer.")
								} else {
									s := []string{urlparams, "&computer=", url.PathEscape(computer)}
									urlparams = strings.Join(s, "")
								}

								if params == "" {
									// fmt.Println("No parameters specified.")
								} else {
									s := []string{urlparams, "&params=", url.PathEscape(params)}
									urlparams = strings.Join(s, "")
								}

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
						}
					}
				}
			}
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
