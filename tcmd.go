package main

import (
        "fmt"
        "io/ioutil"
        "net/http"
        "runtime"
        "os"
        "path/filepath"
        "strings"
)

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
        dir := UserHomeDir()
        fmt.Println(dir)
  
        p := filepath.Join(dir, "/.TRIGGERcmdData/token.tkn")
        fmt.Println(p);

        b, err := ioutil.ReadFile(p) // just pass the file name
        if err != nil {
            fmt.Print(err)
        }

        token := string(b) // convert content to a 'string'

	s := []string{"https://www.triggercmd.com/api/run/triggersave?token=", token, "&computer=russfam&trigger=calculator"}
	fmt.Println(strings.Join(s, ""))

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