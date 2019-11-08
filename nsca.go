package main

import (
	"bytes"
	"strings"
	"os/exec"
	"io/ioutil"
	"os"
	"flag"
    "fmt"
    "log"
	"net/http"
	"github.com/gorilla/mux"
	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/yaml.v2"
)

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        if r.Header["Token"] != nil {

            token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("There was an error")
                }
                return []byte(appConfig.JwtKey), nil
            })

            if err != nil {
                fmt.Fprintf(w, err.Error())
            }

            if token.Valid {
                endpoint(w, r)
            }
        } else {

            fmt.Fprintf(w, "Not Authorized")
        }
    })
}

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World")
    fmt.Println("Endpoint Hit: homePage")
}

func runSystemCommand(command string) (o string) {
	var output strings.Builder
	cmdToken := strings.Split(command, " ")
	var cmd *exec.Cmd
	cmd = exec.Command(cmdToken[0], cmdToken[1:]... )

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Start()
	if err != nil {
		log.Printf("ERROR %v\n", err)
	} else {
		cmd.Wait()
		fmt.Fprintf(&output, "%d\n", cmd.ProcessState.ExitCode())
		fmt.Fprintf(&output, "%s - %s", out.String(), errOut.String())
	}
	return output.String()
}

func ProcessCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	CommandName := vars["CommandName"]
	Commands := appConfig.Commands
	for _, cmd := range(Commands) {
		switch cmd.Name {
		case CommandName:
			ouput := runSystemCommand(cmd.Path)
			w.Write([]byte(ouput))
		}
	}
}

func handleRequests() {
	router := mux.NewRouter()
	router.Handle("/", isAuthorized(homePage)).Methods("GET")
	router.Handle("/run/{CommandName}", isAuthorized(ProcessCommand)).Methods("GET")
	log.Printf("Start server on port %d\n", appConfig.Port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Port), router))
}

type Command struct {
	Name string
	Path string
}

type AppConfig struct { //Why do I have to tag every field! Because yaml driver automatically lowercase the field name to look into the yaml file <yuk>
	Port int
	Commands []Command
	JwtKey string
}

var appConfig AppConfig

func generateDefaultConfig(fPath string) (e error) {
	defaultConfig := `
port: 8000
jwtkey: kGay08Hf5KvSIhYREkiq2FJYNstQsrTK
commands:
  - name: example_ls
    path: /bin/ls
`
	err := ioutil.WriteFile(fPath, []byte(defaultConfig), 0600)
	if err != nil {return err}
	return loadConfig(fPath)
}

func loadConfig(fPath string) (e error) {
	yamlStr, e := ioutil.ReadFile(fPath)
	if e != nil {
		return e
	}
	e = yaml.Unmarshal(yamlStr, &appConfig)
	return e
}

func main() {
	defaultConfig :=  fmt.Sprintf("%s/.nsca-go.yaml", os.Getenv("HOME"))
	configFile := flag.String("c", defaultConfig, fmt.Sprintf("Config file, default %s", defaultConfig) )
	flag.Parse()

	e := loadConfig(*configFile)
	if e != nil {
		log.Printf("Error reading config file. %v\nGenerating new one\n", e)
		if e = generateDefaultConfig(*configFile); e != nil {
			log.Fatalf("ERROR can not geenrate config file %v\n", e)
		}
	}
    handleRequests()
}