package cmd

import (
	"bytes"
	"strings"
	"os/exec"
    "fmt"
    "log"
	"net/http"
	"github.com/gorilla/mux"
	jwt "github.com/dgrijalva/jwt-go"

)

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        if r.Header["Token"] != nil {

            token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("There was an error")
                }
                return []byte(Config.JwtKey), nil
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
	Commands := Config.Commands
	for _, cmd := range(Commands) {
		switch cmd.Name {
		case CommandName:
			ouput := runSystemCommand(cmd.Path)
			w.Write([]byte(ouput))
		}
	}
}

//HandleRequests -
func HandleRequests() {
	router := mux.NewRouter()
	router.Handle("/", isAuthorized(homePage)).Methods("GET")
	router.Handle("/run/{CommandName}", isAuthorized(ProcessCommand)).Methods("GET")
	log.Printf("Start server on port %d\n", Config.Port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), router))
}
//StartServer - We may spawn other listener within this func
func StartServer() {
	HandleRequests()
}