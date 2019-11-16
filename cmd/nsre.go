package cmd

import (
	"strconv"
	"text/template"
	"io/ioutil"
	"bytes"
	"strings"
	"os/exec"
    "fmt"
    "log"
	"net/http"
	"github.com/gorilla/mux"
	jwt "github.com/dgrijalva/jwt-go"
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

//ProcessSearchLog -
func ProcessSearchLog(w http.ResponseWriter, r *http.Request) {
	const tString = `<!DOCTYPE html>
<head>
    <title>webgui</title>
</head>
<body>
    <h1>Search Log</h1>
    <form action="/searchlog" method="POST">
        <table>
            <tr>
                <td><label for="keyword">Keyword: </label></td>
				<td><input name="keyword" id="keyword" type="text" value="{{ .keyword }}" title="keyword to search, understand & to search with AND logic."/></td>
				<td><input type="checkbox" name="sortorder" value="DESC" {{ .sortorder }}>Sort Descending</td>
            </tr>
            <tr>
				<td colspan="2" align="center">
					<input type="button" value="clear" onclick="document.getElementById('keyword').value = '';">&nbsp
					<input name="submit" type="submit" value="submit">
				</td>
            </tr>
    	</table>
	</form>
	<hr/>
	<h2>Result:</h2>
	<p>Found {{ .count }} records</p>
    {{ .output }}

</body>`
	var output strings.Builder
	switch r.Method {
	case "GET":
		t := template.Must(template.New("webgui").Parse(tString))
		e := t.Execute(w, map[string]string{
			"sortorder": "checked",
			"keyword": "",
			})
		if e != nil {
			fmt.Printf("%v\n", e)
		}

	case "POST":
		r.ParseForm()
		keyword := r.FormValue("keyword")
		sortorder := r.Form["sortorder"]
		var sortorderVal, checkedSort string
		if len(sortorder) == 0 {
			sortorderVal = "ASC"
			checkedSort = ""
		} else {
			sortorderVal = "DESC"
			checkedSort = "checked"
		}
		c := SearchLog(keyword, &output, sortorderVal)
		t := template.Must(template.New("webgui").Parse(tString))
		e := t.Execute(w, map[string]string{
			"count": strconv.FormatInt(int64(c), 10),
			"output": output.String(),
			"sortorder": checkedSort,
			"keyword": keyword,
			})
		if e != nil {
			fmt.Printf("%v\n", e)
		}
	}
}

//SearchLog -
func SearchLog(keyword string, o *strings.Builder, sortorder string) (int) {
	q := ""
	keyword = strings.TrimSpace(keyword)
	tokens := strings.Split(keyword, " & ")
	_l := len(tokens)

	for i, t := range(tokens) {
		if i == _l - 1 {
			q = fmt.Sprintf("%s (host LIKE '%%%s%%' OR application LIKE '%%%s%%' OR message LIKE '%%%s%%') ORDER BY timestamp %s LIMIT 200;", q, t, t, t, sortorder)
		} else {
			q = fmt.Sprintf("%s (host LIKE '%%%s%%' OR application LIKE '%%%s%%' OR message LIKE '%%%s%%') AND ", q, t, t, t)
		}
	}
	q = fmt.Sprintf("SELECT timestamp, datelog, host, application, message from log WHERE %s", q)
	fmt.Println(q)

	conn := GetDBConn()
	defer conn.Close()

	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Printf("ERROR - %v\n", err)
	}
	defer stmt.Close()
	fmt.Fprintf(o, "<table>")
	count := 0
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Printf("ERROR - %v\n", err)
		}
		if !hasRow {
			break
		}
		var timestampVal, datelogVal int64
		var host, application, msg string

		err = stmt.Scan(&timestampVal, &datelogVal, &host, & application, &msg)
		if err != nil {
			log.Printf("ERROR - %v\n", err)
		}
		timestamp, datelog := NsToTime(timestampVal), NsToTime(datelogVal)
		AUTimeLayout := "02/01/2006 15:04:05 MST"
		line := fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, timestamp.Format(AUTimeLayout), datelog.Format(AUTimeLayout), template.HTMLEscapeString(host), template.HTMLEscapeString(application), template.HTMLEscapeString(msg))
		fmt.Fprintf(o, line)
		count = count + 1
	}
	fmt.Fprintf(o, "</table>")
	return count
}

//HandleRequests -
func HandleRequests() {
	router := mux.NewRouter()
	router.Handle("/", isAuthorized(homePage)).Methods("GET")
	router.Handle("/run/{CommandName}", isAuthorized(ProcessCommand)).Methods("GET")
	router.Handle("/log", isAuthorized(ProcessLog)).Methods("POST")
	router.HandleFunc("/searchlog", ProcessSearchLog)
	if Config.Sslkey != "" {
		log.Printf("Start SSL/TLS server on port %d\n", Config.Port)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", Config.Port), Config.Sslcert, Config.Sslkey, router))
	} else {
		log.Printf("Start server on port %d\n", Config.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), router))
	}

}
//StartServer - We may spawn other listener within this func
func StartServer() {
	SetUpLogDatabase()
	HandleRequests()
}

//SetUpLogDatabase -
func SetUpLogDatabase() {
	conn := GetDBConn()
	defer conn.Close()

	// err = conn.Exec(`
	// CREATE VIRTUAL TABLE IF NOT EXISTS log USING fts5(timestamp, datelog, host, application, message);
	// PRAGMA main.synchronous=OFF;
	// `)
	err := conn.Exec(`
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
	conn := GetDBConn()
	defer conn.Close()

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	logData := LogData{}
	if e := json.Unmarshal(data, &logData); e != nil {
		log.Printf("ERROR - can not parse json data for logline - %v\n", e)
	}
	err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, message) VALUES (?, ?, ?, ?, ?)`, logData.Timestamp, logData.Datelog, logData.Host, logData.Application, logData.Message)
	if err != nil {
		log.Printf("ERROR - can not insert data for logline - %v\n", err)
	}
}