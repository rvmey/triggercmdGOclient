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
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	var tui bool
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
	app.Version = "1.0.7"
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
		cli.BoolFlag{
			Name:        "tui",
			Usage:       "Launch interactive text user interface",
			Destination: &tui,
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
							err := os.MkdirAll(filepath.Join(dir, "/.TRIGGERcmdData"), 0755)
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
							if tui {
								if err := runTUI(token); err != nil {
									fmt.Println("TUI error:", err)
								}
							} else if trigger == "" {
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

// runTUI launches the interactive text user interface.
func runTUI(token string) error {
	// Fetch commands.
	resp, err := http.Get("https://www.triggercmd.com/api/command/list?token=" + token)
	if err != nil {
		return fmt.Errorf("failed to fetch commands: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	var cmdResult map[string]interface{}
	if err := json.Unmarshal(body, &cmdResult); err != nil {
		return fmt.Errorf("failed to parse commands response: %w", err)
	}
	rawCommands, _ := cmdResult["records"].([]interface{})

	type Command struct {
		Name        string
		Computer    string
		AllowParams bool
		Icon        string
	}
	var allCommands []Command
	computerSet := make(map[string]bool)
	for _, raw := range rawCommands {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		compObj, ok := m["computer"].(map[string]interface{})
		if !ok {
			continue
		}
		compName, _ := compObj["name"].(string)
		cmdName, _ := m["name"].(string)
		allowParams := m["allowParams"] == true
		icon, _ := m["icon"].(string)
		allCommands = append(allCommands, Command{Name: cmdName, Computer: compName, AllowParams: allowParams, Icon: icon})
		computerSet[compName] = true
	}
	computers := make([]string, 0, len(computerSet))
	for c := range computerSet {
		computers = append(computers, c)
	}
	sort.Strings(computers)

	// Fetch panel buttons.
	resp2, err := http.Get("https://triggercmd.com/api/panelbutton/list?token=" + token)
	if err != nil {
		return fmt.Errorf("failed to fetch panels: %w", err)
	}
	defer resp2.Body.Close()
	body2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		return fmt.Errorf("failed to read panels response: %w", err)
	}
	var panelResult map[string]interface{}
	if err := json.Unmarshal(body2, &panelResult); err != nil {
		return fmt.Errorf("failed to parse panels response: %w", err)
	}
	rawButtons, _ := panelResult["records"].([]interface{})

	type PanelButton struct {
		Panel  string
		Button string
		Params string
	}
	var allPanelButtons []PanelButton
	panelSet := make(map[string]bool)
	for _, raw := range rawButtons {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		panelObj, ok := m["panel"].(map[string]interface{})
		if !ok {
			continue
		}
		panelName, _ := panelObj["name"].(string)
		buttonName, _ := m["name"].(string)
		buttonParams, _ := m["params"].(string)
		allPanelButtons = append(allPanelButtons, PanelButton{Panel: panelName, Button: buttonName, Params: buttonParams})
		panelSet[panelName] = true
	}
	panels := make([]string, 0, len(panelSet))
	for p := range panelSet {
		panels = append(panels, p)
	}
	sort.Strings(panels)

	// --- Build TUI ---
	app := tview.NewApplication()
	pages := tview.NewPages()

	// Header bar showing active mode.
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)

	// Shared result panel.
	statusView := tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	statusView.SetTitle(" Result ").SetBorder(true)

	// Commands mode widgets.
	computerList := tview.NewList().ShowSecondaryText(false)
	computerList.SetTitle(" Computers ").SetBorder(true)

	commandList := tview.NewList().ShowSecondaryText(false)
	commandList.SetTitle(" Commands ").SetBorder(true)

	// Panels mode widgets.
	panelList := tview.NewList().ShowSecondaryText(false)
	panelList.SetTitle(" Panels ").SetBorder(true)

	buttonList := tview.NewList().ShowSecondaryText(false)
	buttonList.SetTitle(" Buttons ").SetBorder(true)

	// Left column toggles between computer list and panel list.
	leftPages := tview.NewPages()
	leftPages.AddPage("computers", computerList, true, true)
	leftPages.AddPage("panels", panelList, true, false)

	// Right list area toggles between command list and button list.
	rightListPages := tview.NewPages()
	rightListPages.AddPage("commands", commandList, true, true)
	rightListPages.AddPage("panels", buttonList, true, false)

	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(rightListPages, 0, 3, false).
		AddItem(statusView, 6, 1, false)

	bodyFlex := tview.NewFlex().
		AddItem(leftPages, 30, 0, true).
		AddItem(rightFlex, 0, 1, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(bodyFlex, 0, 1, true)

	pages.AddPage("main", mainFlex, true, true)

	// triggerCommand calls the API to run a command and shows the result.
	triggerCommand := func(computerName, triggerName, params string) {
		urlStr := "https://www.triggercmd.com/api/run/triggersave?token=" + token +
			"&trigger=" + url.QueryEscape(triggerName) +
			"&computer=" + url.QueryEscape(computerName)
		if params != "" {
			urlStr += "&params=" + url.QueryEscape(params)
		}
		r, err := http.Get(urlStr)
		if err != nil {
			statusView.SetText("[red]Error: " + err.Error())
			return
		}
		defer r.Body.Close()
		b, _ := ioutil.ReadAll(r.Body)
		statusView.SetText(string(b))
	}

	// triggerPanel calls the API to trigger a panel button.
	triggerPanel := func(panelName, buttonName, params string) {
		urlStr := "https://www.triggercmd.com/api/panel/trigger?token=" + token +
			"&panel=" + url.QueryEscape(panelName) +
			"&button=" + url.QueryEscape(buttonName)
		if params != "" {
			urlStr += "&params=" + url.QueryEscape(params)
		}
		r, err := http.Get(urlStr)
		if err != nil {
			statusView.SetText("[red]Error: " + err.Error())
			return
		}
		defer r.Body.Close()
		b, _ := ioutil.ReadAll(r.Body)
		statusView.SetText(string(b))
	}

	// showParamModal presents an input form for commands that accept parameters.
	showParamModal := func(computerName, triggerName string) {
		var paramsInput string

		form := tview.NewForm()
		form.AddInputField("Parameters:", "", 46, nil, func(text string) {
			paramsInput = text
		})
		form.AddButton("Run", func() {
			pages.RemovePage("params")
			app.SetFocus(commandList)
			triggerCommand(computerName, triggerName, paramsInput)
		})
		form.AddButton("Cancel", func() {
			pages.RemovePage("params")
			app.SetFocus(commandList)
		})
		form.SetTitle(" Parameters for \"" + triggerName + "\" ").SetBorder(true)
		form.SetCancelFunc(func() {
			pages.RemovePage("params")
			app.SetFocus(commandList)
		})

		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 9, 1, true).
				AddItem(nil, 0, 1, false), 60, 1, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("params", modal, true, true)
		app.SetFocus(form)
	}

	// updateCommandList repopulates the command list for the chosen computer.
	updateCommandList := func(computerName string) {
		commandList.Clear()
		statusView.SetText("[grey]Select a command to run it.")
		for _, cmd := range allCommands {
			if cmd.Computer != computerName {
				continue
			}
			c := cmd
			label := c.Name
			if c.Icon != "" {
				label = c.Icon + " " + c.Name
			}
			if c.AllowParams {
				label = label + "  [grey](accepts parameters)[-]"
			}
			commandList.AddItem(label, "", 0, func() {
				if c.AllowParams {
					showParamModal(c.Computer, c.Name)
				} else {
					triggerCommand(c.Computer, c.Name, "")
				}
			})
		}
		app.SetFocus(commandList)
	}

	// showParamSelectModal presents a list of choices when params is comma-separated.
	showParamSelectModal := func(panelName, buttonName, paramsList string) {
		choices := strings.Split(paramsList, ",")
		for i, c := range choices {
			choices[i] = strings.TrimSpace(c)
		}

		optionList := tview.NewList().ShowSecondaryText(false)
		optionList.SetTitle(" Select parameter ").SetBorder(true)
		for _, choice := range choices {
			v := choice
			optionList.AddItem(v, "", 0, func() {
				pages.RemovePage("paramsel")
				app.SetFocus(buttonList)
				triggerPanel(panelName, buttonName, v)
			})
		}
		optionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.RemovePage("paramsel")
				app.SetFocus(buttonList)
				return nil
			}
			return event
		})

		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(optionList, len(choices)+4, 1, true).
				AddItem(nil, 0, 1, false), 40, 1, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("paramsel", modal, true, true)
		app.SetFocus(optionList)
	}

	// updateButtonList repopulates the button list for the chosen panel.
	updateButtonList := func(panelName string) {
		buttonList.Clear()
		statusView.SetText("[grey]Select a button to trigger it.")
		for _, pb := range allPanelButtons {
			if pb.Panel != panelName {
				continue
			}
			b := pb
			label := b.Button
			if strings.Contains(b.Params, ",") {
				label = b.Button + "  [grey](" + b.Params + ")[-]"
			}
			buttonList.AddItem(label, "", 0, func() {
				if strings.Contains(b.Params, ",") {
					showParamSelectModal(b.Panel, b.Button, b.Params)
				} else {
					triggerPanel(b.Panel, b.Button, "")
				}
			})
		}
		app.SetFocus(buttonList)
	}

	// setMode switches the TUI between commands and panels modes.
	setMode := func(m string) {
		if m == "panels" {
			leftPages.SwitchToPage("panels")
			rightListPages.SwitchToPage("panels")
			header.SetText("  F1 Commands   [::b][F2 Panels][-]   Esc: quit / back")
			statusView.SetText("[grey]Select a panel, then select a button to trigger it.")
			app.SetFocus(panelList)
		} else {
			leftPages.SwitchToPage("computers")
			rightListPages.SwitchToPage("commands")
			header.SetText("  [::b][F1 Commands][-]   F2 Panels   Esc: quit / back")
			statusView.SetText("[grey]Select a computer, then select a command to run it.")
			app.SetFocus(computerList)
		}
	}

	// Populate computer list.
	for _, name := range computers {
		n := name
		computerList.AddItem(n, "", 0, func() {
			updateCommandList(n)
		})
	}

	// Populate panel list.
	for _, name := range panels {
		n := name
		panelList.AddItem(n, "", 0, func() {
			updateButtonList(n)
		})
	}

	// Key bindings for computer list.
	computerList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			if commandList.GetItemCount() > 0 {
				app.SetFocus(commandList)
			}
			return nil
		case tcell.KeyF2:
			setMode("panels")
			return nil
		case tcell.KeyEscape:
			app.Stop()
			return nil
		}
		return event
	})

	// Key bindings for command list.
	commandList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab, tcell.KeyLeft:
			app.SetFocus(computerList)
			return nil
		case tcell.KeyF2:
			setMode("panels")
			return nil
		case tcell.KeyEscape:
			app.SetFocus(computerList)
			return nil
		}
		return event
	})

	// Key bindings for panel list.
	panelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			if buttonList.GetItemCount() > 0 {
				app.SetFocus(buttonList)
			}
			return nil
		case tcell.KeyF1:
			setMode("commands")
			return nil
		case tcell.KeyEscape:
			app.Stop()
			return nil
		}
		return event
	})

	// Key bindings for button list.
	buttonList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab, tcell.KeyLeft:
			app.SetFocus(panelList)
			return nil
		case tcell.KeyF1:
			setMode("commands")
			return nil
		case tcell.KeyEscape:
			app.SetFocus(panelList)
			return nil
		}
		return event
	})

	// Start in commands mode; pre-populate both sides.
	if len(computers) > 0 {
		updateCommandList(computers[0])
	}
	if len(panels) > 0 {
		updateButtonList(panels[0])
	}
	setMode("commands")

	return app.SetRoot(pages, true).EnableMouse(true).Run()
}
