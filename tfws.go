package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

var termState *term.State
var workspaces []string

func saveTermState() {
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	termState = oldState
}

func restoreTermState() {
	if termState != nil {
		defer term.Restore(int(os.Stdin.Fd()), termState)
	}
}

func getWorkspaces() []string {
	yfile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		log.Fatal(err)
		//restoreTermState()
	}

	data := make(map[interface{}]interface{})

	err2 := yaml.Unmarshal(yfile, &data)

	if err2 != nil {
		log.Fatal(err2)
	}

	wsList := []string{}
	for k, _ := range data {
		wsList = append(wsList, k.(string))
	}
	return wsList
}

func wsOptions(input prompt.Document) []prompt.Suggest {
	suggests := []prompt.Suggest{}
	for _, command := range workspaces {
		//fmt.Println(command)
		suggests = append(suggests, prompt.Suggest{
			Text:        string(command),
			Description: "",
		})
	}
	return prompt.FilterHasPrefix(suggests, input.GetWordBeforeCursor(), true)
}

func runCommand(input string) (result *exec.Cmd) {
	command := strings.Split(input, " ")
	cmd := exec.Command(command[0], command[1:]...)
	var out strings.Builder
	var outErr strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err := cmd.Run()

	if err != nil {
		fmt.Printf("%s", outErr.String())
	}

	return cmd
}

func main() {
	saveTermState()
	workspaces = getWorkspaces()

	fmt.Println("Workspaces")
	e := prompt.Input("> ", wsOptions)

	if len(workspaces) > 1 {
		r := runCommand("terraform workspace select " + e)
		if r.Stderr != nil {
			fmt.Println("Creating new workspace")
			r := runCommand("terraform workspace new " + e)
			if r.Err != nil {
				fmt.Printf("%s", r.String())
			}
			fmt.Printf("%s", r.Stdout)
		}
		fmt.Printf("%s", r.Stdout)

	} else {
		fmt.Println("No other workspaces other than default found")
	}

	restoreTermState()
}
