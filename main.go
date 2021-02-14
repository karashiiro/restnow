package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var fileTemplate = `package main

import (
	"flag"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := flag.Uint("port", {{.Port}}, "server binding port")
	flag.Parse()

	app := fiber.New()

	{{.BuiltRoutes}}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
`

var methodTemplate = `
	app.{{.Method}}("{{.Route}}", func(ctx *fiber.Ctx) error {
		return nil
	})
`

// MethodInfo represents template filler for the method info template.
type MethodInfo struct {
	Method string
	Route  string
}

// ProjectConfiguration represents the schema of the json file used to set up the project.
type ProjectConfiguration struct {
	Name     string                 `json:"name"`
	RepoName string                 `json:"repoName"`
	Port     uint                   `json:"defaultPort"`
	Routes   map[string]interface{} `json:"routes"`

	BuiltRoutes string
}

// FirstOrDefault clones similar LINQ functionality.
func FirstOrDefault(arr []string, predicate func(string) bool) string {
	for _, s := range arr {
		if predicate(s) {
			return s
		}
	}
	return ""
}

// Fatalln is a clone of log.Fatalln that doesn't include a timestamp.
func Fatalln(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

// RunCommand runs the specified command in the provided working directory.
func RunCommand(command string, cwd string) error {
	args := strings.Split(command, " ")

	initCmd := exec.Command(args[0], args[1:]...)

	dir, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}
	initCmd.Dir = dir

	err = initCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// BuildRoutes builds the router functions from the route configuration.
func BuildRoutes(routes map[string]interface{}, curRoute string) (string, error) {
	finalOut := ""
	for route, children := range routes {
		// Cast to method array
		if cs, ok := children.([]interface{}); ok {
			for _, c := range cs {
				// Cast method interface{} to string
				m, ok := c.(string)
				if !ok {
					fmt.Println(c)
					return "", errors.New("error: unknown struct")
				}

				// Format method to Fiber method name
				method := strings.ToLower(m)
				method = strings.ToUpper(string(method[0])) + method[1:]

				methodInfo := &MethodInfo{Method: method, Route: curRoute + "/" + route}

				tmpl, err := template.New(route + m).Parse(methodTemplate)
				if err != nil {
					return "", err
				}

				outTxtStr := ""
				outTxt := bytes.NewBufferString(outTxtStr)

				err = tmpl.Execute(outTxt, methodInfo)
				if err != nil {
					return "", err
				}

				finalOut += outTxt.String() + "\n"
			}
		} else if nextRoutes, ok := children.(map[string]interface{}); ok {
			nextOut, err := BuildRoutes(nextRoutes, curRoute+"/"+route)
			if err != nil {
				return "", err
			}
			finalOut += nextOut
		} else {
			fmt.Println(children)
			return "", errors.New("error: unknown struct")
		}
	}
	return finalOut, nil
}

func main() {
	jsonPath := FirstOrDefault(os.Args[1:], func(s string) bool {
		_, err := os.Stat(s)
		return err == nil
	})

	jsonData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		Fatalln(err)
	}

	projectConfig := &ProjectConfiguration{}
	if err := json.Unmarshal(jsonData, projectConfig); err != nil {
		Fatalln(err)
	}

	if _, err = os.Stat(projectConfig.Name); err == nil {
		Fatalln("Project folder already exists!")
	}

	os.Mkdir(projectConfig.Name, 0755)

	out, err := template.New("main").Parse(fileTemplate)
	if err != nil {
		Fatalln(err)
	}

	outFile, err := os.Create(path.Join(projectConfig.Name, "main.go"))
	if err != nil {
		Fatalln(err)
	}
	defer outFile.Close()

	projectConfig.BuiltRoutes, err = BuildRoutes(projectConfig.Routes, "")
	if err != nil {
		Fatalln(err)
	}

	err = out.Execute(outFile, projectConfig)
	if err != nil {
		Fatalln(err)
	}

	commands := []string{"git init", "go mod init " + projectConfig.RepoName, "go get -d ./...", "go fmt main.go"}
	for _, cmd := range commands {
		err = RunCommand(cmd, projectConfig.Name)
		if err != nil {
			Fatalln(err)
		}
	}
}
