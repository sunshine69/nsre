package cmd

import (
	"io/ioutil"
	"time"
	"bytes"
	"strings"
	"os/exec"
    "fmt"
    "log"
	"net/http"
	"github.com/gorilla/mux"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/json-iterator/go"
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

func ProcessLog(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	InsertLog(body)
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
	router.Handle("/log", isAuthorized(ProcessLog)).Methods("POST")
	log.Printf("Start server on port %d\n", Config.Port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), router))
}
//StartServer - We may spawn other listener within this func
func StartServer() {
	SetUpLogDatabase()
	HandleRequests()
}

//SetUpLogDatabase -
func SetUpLogDatabase() {
	conn, err := sqlite3.Open(Config.Logdbpath)
	if err != nil {
		log.Fatalf("ERROR - can not open log database file - %v\n", err)
	}
	defer conn.Close()
	conn.BusyTimeout(5 * time.Second)
	// err = conn.Exec(`
	// CREATE VIRTUAL TABLE IF NOT EXISTS log USING fts5(timestamp, datelog, host, application, message);
	// PRAGMA main.synchronous=OFF;
	// `)
	err = conn.Exec(`
	CREATE TABLE IF NOT EXISTS log(id integer primary key autoincrement,timestamp int, datelog int, host text, application text, message text);
	CREATE UNIQUE INDEX IF NOT EXISTS t_host_idx ON log(timestamp, host);
	PRAGMA main.page_size = 4096;
	PRAGMA main.cache_size=10000;
	PRAGMA main.locking_mode=EXCLUSIVE;
	PRAGMA main.synchronous=NORMAL;
	PRAGMA main.journal_mode=WAL;
	PRAGMA main.cache_size=5000;
	`)
	if err != nil {
		log.Fatalf("ERROR - can not create table log - %v\n", err)
	}
}

//LogData -
type LogData struct {
	Timestamp int64
	Datelog int64
	Host string
	Application string
	Message string
}

//InsertLog -
func InsertLog(data []byte) {
	conn, err := sqlite3.Open(Config.Logdbpath)
	if err != nil {
		log.Fatalf("ERROR - can not open log database file - %v\n", err)
	}
	defer conn.Close()

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	logData := LogData{}
	if e := json.Unmarshal(data, &logData); e != nil {
		log.Printf("ERROR - can not parse json data for logline - %v\n", e)
	}
	err = conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, message) VALUES (?, ?, ?, ?, ?)`, logData.Timestamp, logData.Datelog, logData.Host, logData.Application, logData.Message)
	if err != nil {
		log.Printf("ERROR - can not insert data for logline - %v\n", err)
	}
}