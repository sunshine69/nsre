package cmd

import (
	"regexp"
	"time"
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

func isOauthAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := SessionStore.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		useremail := session.Values["useremail"]

		if useremail == nil {
			log.Printf("ERROR - No session\n")
			// fmt.Fprintf(w, `<!DOCTYPE html>
			// <html><body>Not Authorized - <a href="/auth/google/login">Login</a></body></html>`)
			http.Redirect(w, r, "/auth/google/login", http.StatusTemporaryRedirect)
			return
		}
		endpoint(w, r)
    })
}

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello. If you see this message, the jwt auth works")
    fmt.Println("Endpoint Hit: homePage. ")
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
		if cmd.ProcessState.ExitCode() != 0 {
			fmt.Fprintf(&output, "%s - %s", out.String(), errOut.String())
		} else {
			fmt.Fprintf(&output, "%s", out.String())
		}
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
			output := runSystemCommand(cmd.Path)
			w.Write([]byte(output))
		default:
			output := "2\nERROR - Command "+CommandName +" does not exists"
			w.Write([]byte(output))
		}
	}
}

//ProcessSearchLogByID - Take an ID and search for record surrounding including the current rec with time span of 10 minutes
func ProcessSearchLogByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session, _ := SessionStore.Get(r, "auth-session")
	var sortorder, duration, tz interface{}
	if sortorder = session.Values["sortorder"]; sortorder == nil {
		sortorder = "checked"
		session.Values["sortorder"] = sortorder.(string)
	}
	if duration = session.Values["duration"]; duration == nil {
		duration = "15m"
		session.Values["duration"] = duration.(string)
	}
	if tz = session.Values["tz"]; tz == nil {
		tz = "AEST"
		session.Values["tz"] = tz.(string)
	}
	id := vars["id"]
	q := `SELECT timestamp, host, application from log WHERE id = '` + id + `'`
	conn := GetDBConn()
	defer conn.Close()
	stmt, _ := conn.Prepare(q)
	defer stmt.Close()
	var timestamp int64
	var host, application string
	hasRow, _ := stmt.Step()
	if !hasRow {
		fmt.Fprintf(w, "No row found.")
		return
	}
	stmt.Scan(&timestamp, &host, &application)
	rowTime := NsToTime(timestamp)
	//Take the duration from the form input and use it as timerange
	_start, _end := ParseTimeRange(duration.(string), tz.(string))
	halfDuration := _end.Sub(_start) / 2

	start := rowTime.Add(-1 * halfDuration)
	end := rowTime.Add(halfDuration)
	log.Printf("Time range: %s - %s\n",start.Format(AUTimeLayout), end.Format(AUTimeLayout)  )
	q = fmt.Sprintf("SELECT id, timestamp, datelog, host, application, logfile, message from log WHERE ((timestamp > %d) AND (timestamp < %d)) AND host = '%s' AND application = '%s'", start.UnixNano(), end.UnixNano(), host, application)

	var output strings.Builder
	c := DoSQLSearch(q, &output)
	//Load template and call search function
	tStringb, _ := ioutil.ReadFile("templates/searchpage.go.html")
	tString := string(tStringb)

	session.Save(r, w)

	t := template.Must(template.New("webgui").Parse(tString))
	e := t.Execute(w, map[string]string{
		"count": strconv.FormatInt(int64(c), 10),
		"output": output.String(),
		"sortorder": sortorder.(string),
		"keyword": "",
		"duration": duration.(string),
		"tz": tz.(string),
		})
	if e != nil {
		fmt.Printf("%v\n", e)
	}
}

//ProcessSearchLog -
func ProcessSearchLog(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "auth-session")
	tStringb, _ := ioutil.ReadFile("templates/searchpage.go.html")
	tString := string(tStringb)
	var output strings.Builder

	switch r.Method {
	case "GET":
		var sortorder, duration, tz interface{}
		if sortorder = session.Values["sortorder"]; sortorder == nil {
			sortorder = "checked"
			session.Values["sortorder"] = sortorder.(string)
		}
		if duration = session.Values["duration"]; duration == nil {
			duration = "15m"
			session.Values["duration"] = duration.(string)
		}
		if tz = session.Values["tz"]; tz == nil {
			tz = "AEST"
			session.Values["tz"] = tz.(string)
		}
		session.Save(r, w)

		t := template.Must(template.New("webgui").Parse(tString))
		e := t.Execute(w, map[string]string{
			"sortorder": sortorder.(string),
			"keyword": "",
			"duration": duration.(string),
			"tz": tz.(string),
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
			session.Values["sortorder"] = checkedSort
		} else {
			sortorderVal = "DESC"
			checkedSort = "checked"
			session.Values["sortorder"] = checkedSort
		}
		duration := r.FormValue("duration")
		tz := r.FormValue("tz")
		session.Values["duration"] = duration
		session.Values["tz"] = tz
		session.Save(r,w)

		c := SearchLog(keyword, &output, sortorderVal, duration, tz)
		t := template.Must(template.New("webgui").Parse(tString))
		e := t.Execute(w, map[string]string{
			"count": strconv.FormatInt(int64(c), 10),
			"output": output.String(),
			"sortorder": checkedSort,
			"keyword": keyword,
			"duration": duration,
			"tz": tz,
			})
		if e != nil {
			fmt.Printf("%v\n", e)
		}
	}
}

//DoSQLSearch - Execute the search in the database. Return the record counts and fill the string builder object.
func DoSQLSearch(q string, o *strings.Builder) (int) {
	fmt.Println(q)

	conn := GetDBConn()
	defer conn.Close()

	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Printf("ERROR - %v\n", err)
	}
	defer stmt.Close()
	fmt.Fprintf(o, `
	<table id="customers">
		<col width="10%">
		<col width="10%">
		<col width="10%">
		<col width="10%">
		<col width="60%">
		<tr>
			<th>TS</th>
			<th>Date</th>
			<th>Host</th>
			<th>Application</th>
			<th>Message</th>
		</tr>
	`)
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
		var id int
		var host, application, logfile, msg string

		err = stmt.Scan(&id, &timestampVal, &datelogVal, &host, & application, &logfile, &msg)
		if err != nil {
			log.Printf("ERROR - %v\n", err)
		}
		timestamp, datelog := NsToTime(timestampVal), NsToTime(datelogVal)

		line := fmt.Sprintf(`
		<tr title="%s">
			<td title="filter similar records around this time"><a href="/searchlogbyid/%d">%s</a></td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, logfile, id, timestamp.Format(AUTimeLayout), datelog.Format(AUTimeLayout), template.HTMLEscapeString(host), template.HTMLEscapeString(application), template.HTMLEscapeString(msg))
		fmt.Fprintf(o, line)
		count = count + 1
	}
	fmt.Fprintf(o, "</table>")
	return count
}

//ParseTimeRange -
func ParseTimeRange(durationStr, tz string) (time.Time, time.Time) {
	var start, end time.Time
	timerangePtn := regexp.MustCompile(`([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2}) - ([\d]{2,2}/[\d]{2,2}/[\d]{4,4} [\d]{2,2}:[\d]{2,2}:[\d]{2,2})`)
	dur, e := time.ParseDuration(durationStr)
	if e != nil {
		m := timerangePtn.FindStringSubmatch(durationStr)
		if len(m) != 3 {
			log.Printf("ERROR Can not parse duration. Set default to 15m ago - %v", e)
			dur, _ = time.ParseDuration("15m")
		} else {
			start, _ = time.Parse(AUTimeLayout, m[1] + " " + tz )
			end, _ = time.Parse(AUTimeLayout, m[2] + " " + tz)
		}
	} else {
		end = time.Now()
		start = end.Add(-1 * dur)
	}
	log.Printf("Time range: %s - %s\n",start.Format(AUTimeLayout), end.Format(AUTimeLayout))
	return start, end
}

//SearchLog -
func SearchLog(keyword string, o *strings.Builder, sortorder, duration, tz string) (int) {
	keyword = strings.TrimSpace(keyword)
	tokens := strings.Split(keyword, " & ")
	_l := len(tokens)

	start, end := ParseTimeRange(duration, tz)

	q := fmt.Sprintf("SELECT id, timestamp, datelog, host, application, logfile, message from log WHERE ((timestamp > %d) AND (timestamp < %d)) AND ", start.UnixNano(), end.UnixNano())

	for i, t := range(tokens) {
		if i == _l - 1 {
			q = fmt.Sprintf("%s (host LIKE '%%%s%%' OR application LIKE '%%%s%%' OR logfile LIKE '%%%s%%' OR message LIKE '%%%s%%') ORDER BY timestamp %s;", q, t, t, t, t, sortorder)
		} else {
			q = fmt.Sprintf("%s (host LIKE '%%%s%%' OR application LIKE '%%%s%%' OR logfile LIKE '%%%s%%' OR message LIKE '%%%s%%') AND ", q, t, t, t, t)
		}
	}
	return DoSQLSearch(q, o)
}

//HandleRequests -
func HandleRequests() {
	router := mux.NewRouter()
	router.HandleFunc("/auth/google/login", OauthGoogleLogin)
	router.HandleFunc("/auth/google/callback", OauthGoogleCallback)
	router.Handle("/searchlog", isOauthAuthorized(ProcessSearchLog))
	router.Handle("/searchlogbyid/{id:[0-9]+}", isOauthAuthorized(ProcessSearchLogByID))

	router.Handle("/", isAuthorized(homePage)).Methods("GET")
	router.Handle("/run/{CommandName}", isAuthorized(ProcessCommand)).Methods("GET")

	if Config.JwtKey == "" {
		log.Printf("WARNING WARNING - JWTKEY is not set. Log server will allow anyone to put log in\n")
		router.HandleFunc("/log/{idx_name}/{type_name}/{unique_id}", ProcessLog).Methods("POST")
		router.HandleFunc("/log/{idx_name}/{type_name}", ProcessLog).Methods("POST")
		router.HandleFunc("/log", ProcessLog).Methods("POST")
	} else{
		router.Handle("/log/{idx_name}/{type_name}/{unique_id}", isAuthorized(ProcessLog)).Methods("POST")
		router.Handle("/log/{idx_name}/{type_name}", isAuthorized(ProcessLog)).Methods("POST")
		router.Handle("/log", isAuthorized(ProcessLog)).Methods("POST")
	}

	srv := &http.Server{
        Addr:  fmt.Sprintf(":%d", Config.Port),
        // Good practice to set timeouts to avoid Slowloris attacks.
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        IdleTimeout:  time.Second * 60,
        Handler: router, // Pass our instance of gorilla/mux in.
    }

	if Config.Sslkey != "" {
		log.Printf("Start SSL/TLS server on port %d\n", Config.Port)
		log.Fatal(srv.ListenAndServeTLS(Config.Sslcert, Config.Sslkey))
	} else {
		log.Printf("Start server on port %d\n", Config.Port)
		log.Fatal(srv.ListenAndServe())
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
	CREATE TABLE IF NOT EXISTS log(id integer primary key autoincrement,timestamp int, datelog int, host text, application text, logfile text, message text);
	CREATE TABLE IF NOT EXISTS user(id integer primary key autoincrement, username text, email text UNIQUE);
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
	Logfile string
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
	err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, logfile, message) VALUES (?, ?, ?, ?, ?, ?)`, logData.Timestamp, logData.Datelog, logData.Host, logData.Application, logData.Logfile, logData.Message)
	if err != nil {
		log.Printf("ERROR - can not insert data for logline - %v\n", err)
	}
}