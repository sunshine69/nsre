package cmd

import (
	"net/http/httputil"
	"mime/multipart"
	"crypto/subtle"
	"bufio"
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

func IsBasicAuth(endpoint func(http.ResponseWriter, *http.Request), username, password, realm string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        user, pass, ok := r.BasicAuth()

        if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
            w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
            w.WriteHeader(401)
            w.Write([]byte("Unauthorised.\n"))
            return
        }
        endpoint(w, r)
    })
}

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
		uri := r.RequestURI
		uri = strings.TrimPrefix(uri, "/")
		session.Values["current_uri"] = uri
		session.Save(r, w)

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
	foundCmd := false
	loop:
	for _, cmd := range(Commands) {
		switch cmd.Name {
		case CommandName:
			output := runSystemCommand(cmd.Path)
			w.Write([]byte(output))
			foundCmd = true
			break loop
		}
	}
	if ! foundCmd {
		output := "2\nERROR - Command "+ CommandName +" does not exists"
		w.Write([]byte(output))
	}
}

//ProcessSearchLogByID - Take an ID and search for record surrounding including the current rec with time span of 10 minutes
func ProcessSearchLogByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session, e := SessionStore.Get(r, "auth-session")
	if e != nil {
		log.Fatalf("Can not get session - %v\n", e)
	}
	var sortorder, duration, tz interface{}
	sortorder = session.Values["sortorder"]
	if sortorder == nil {
		sortorder = "unchecked"
	}

	duration = session.Values["duration"]
	if duration == nil {
		duration = "15m"
		session.Values["duration"] = duration.(string)
	}

	tz = session.Values["tz"]
	if tz == nil {
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

	//Limit the filter time range to maximum 2 hours
	if halfDuration.Hours() > 1 {
		halfDuration, _ = time.ParseDuration("1h")
	}

	start := rowTime.Add(-1 * halfDuration)
	end := rowTime.Add(halfDuration)
	log.Printf("Time range: %s - %s\n",start.Format(AUTimeLayout), end.Format(AUTimeLayout)  )
	q = fmt.Sprintf("SELECT id, timestamp, datelog, host, application, logfile, message from log WHERE ((timestamp > %d) AND (timestamp < %d)) AND host = '%s' AND application = '%s' ORDER BY timestamp ASC", start.UnixNano(), end.UnixNano(), host, application)

	var output strings.Builder
	c := DoSQLSearch(q, &output)

	tString := LoadTemplate("templates/searchpage.go.html")

	t := template.Must(template.New("webgui").Parse(tString))

	if e := session.Save(r, w); e != nil {
        http.Error(w, e.Error(), http.StatusInternalServerError)
        return
    }
	e = t.Execute(w, map[string]string{
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

//LoadTemplate - Run this command to create the bindata.go
// go-bindata -fs -pkg cmd -o cmd/bindata.go -nomemcopy templates
// go get -u github.com/go-bindata/go-bindata/...
func LoadTemplate(tFilePath string) (string) {
	tStringb, e := Asset(tFilePath)
	if e != nil {
		log.Fatalf("ERROR - Can not load template %s - %v\n", tFilePath, e)
	}
	return string(tStringb)
}

//ProcessSearchLog -
func ProcessSearchLog(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "auth-session")

	tString := LoadTemplate("templates/searchpage.go.html")
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

		t := template.Must(template.New("webgui").Parse(tString))
		if e := session.Save(r, w); e !=nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		if e := t.Execute(w, map[string]string{
			"sortorder": sortorder.(string),
			"keyword": "",
			"duration": duration.(string),
			"tz": tz.(string),
			}); e != nil {
			fmt.Printf("%v\n", e)
		}

	case "POST":
		r.ParseForm()
		keyword := r.FormValue("keyword")
		sortorder := r.Form["sortorder"]
		var sortorderVal, checkedSort string
		if len(sortorder) == 0 {
			sortorderVal = "ASC"
			checkedSort = "unchecked"
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
		if e := session.Save(r, w); e !=nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		isSaveValues := r.Form["save_values"]
		fmt.Printf("%v\n", isSaveValues)
		if isSaveValues != nil {
			fmt.Fprintf(w, "Values saved. Click Back to return")
			return
		}

		c := SearchLog(keyword, &output, sortorderVal, duration, tz)
		t := template.Must(template.New("webgui").Parse(tString))

		if e := t.Execute(w, map[string]string{
			"count": strconv.FormatInt(int64(c), 10),
			"output": output.String(),
			"sortorder": checkedSort,
			"keyword": keyword,
			"duration": duration,
			"tz": tz,
			}); e != nil {
			fmt.Printf("%v\n", e)
		}
	}
}

//DoSQLSearch - Execute the search in the database. Return the record counts and fill the string builder object.
func DoSQLSearch(q string, o *strings.Builder) (int) {
	log.Printf("DEBUG - Query '%s'\n", q)

	conn := GetDBConn()
	defer conn.Close()

	stmt, err := conn.Prepare(q)
	if err != nil {
		log.Printf("ERROR - %v\n", err)
	}
	defer stmt.Close()
	fmt.Fprintf(o, `
	<table id="customers">
		<col width="10%%">
		<col width="10%%">
		<col width="10%%">
		<col width="10%%">
		<col width="60%%">
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

//SearchLog -
func SearchLog(keyword string, o *strings.Builder, sortorder, duration, tz string) (int) {
	keyword = strings.TrimSpace(keyword)
	start, end := ParseTimeRange(duration, tz)
	var q string
	if strings.HasPrefix(keyword, "select") || strings.HasPrefix(keyword, "SELECT") {
		timerange := fmt.Sprintf(" WHERE ((timestamp > %d) AND (timestamp < %d)) AND ", start.UnixNano(), end.UnixNano())
		q = strings.Replace(keyword, " WHERE ", timerange, 1) + " ORDER BY timestamp " + sortorder + ";"
		q = strings.Replace(keyword, " where ", timerange, 1) + " ORDER BY timestamp " + sortorder + ";"
	} else {
		splitPtn := regexp.MustCompile(`[\s]+[\&\+][\s]+`)
		// tokens := strings.Split(keyword, " & ")
		tokens := splitPtn.Split(keyword, -1)
		_l := len(tokens)

		q = fmt.Sprintf("SELECT id, timestamp, datelog, host, application, logfile, message from log WHERE ((timestamp > %d) AND (timestamp < %d)) AND ", start.UnixNano(), end.UnixNano())

		for i, t := range(tokens) {
			negate := ""
			combind := "OR"
			if strings.HasPrefix(t, "!") || strings.HasPrefix(t, "-") {
				negate = " NOT "
				t = strings.Replace(t, "!", "", 1)
				t = strings.Replace(t, "-", "", 1)
				combind = "AND"
			}
			if i == _l - 1 {
				q = fmt.Sprintf("%s (host %s LIKE '%%%s%%' %s application %s LIKE '%%%s%%' %s logfile %s LIKE '%%%s%%' %s message %s LIKE '%%%s%%') ORDER BY timestamp %s;", q, negate, t, combind, negate, t, combind, negate, t, combind, negate, t, sortorder)
			} else {
				q = fmt.Sprintf("%s (host %s LIKE '%%%s%%' %s application %s LIKE '%%%s%%' %s logfile %s LIKE '%%%s%%' %s message %s LIKE '%%%s%%') AND ", q, negate,t, combind, negate, t, combind, negate, t, combind, negate,t)
			}
		}
	}
	return DoSQLSearch(q, o)
}

func SendProcessCommand(w http.ResponseWriter, r *http.Request) {
	userIPWithPort := ReadUserIP(r)
	userIP := strings.Split(userIPWithPort, ":")[0]
	fmt.Printf("DEBUG userID %s\n", userIP)
	if userIP != "127.0.0.1" {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	proto := vars["proto"]
	port := vars["port"]
	remoteNsreHost := vars["remote_nsre_host"]
	remoteNsreURL := proto + "://" + remoteNsreHost + ":" + port
	CommandName := vars["CommandName"]
	//Send comand to remote
	o := RunCommand(CommandName, "quiet", remoteNsreURL)
	fmt.Fprintf(w, o)
	return
}

const MaxUploadSizeInMemory = 4 * 1024 * 1024 // 4 MB
const MaxUploadSize = 4 * 1024 * 1024 * 1024
//Present a simple form to allow user to create a log entry. Or upload a text file and parse log
func DoUploadLog(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tString := LoadTemplate("templates/load_log_form.go.html")
		t := template.Must(template.New("load_log_form").Parse(tString))
		t.Execute(w, map[string]interface{}{
		})
		return
	case "POST":
		if err := r.ParseMultipartForm(MaxUploadSizeInMemory); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		host := r.FormValue("host")
		application := r.FormValue("application")
		logfile := r.FormValue("logfile")
		message := r.FormValue("message")
		conn := GetDBConn()
		defer conn.Close()

		var newIDList []int64

		if message != ""{
			message = FilterPassword(message, PasswordFilterPtns)
			message = DecodeJenkinsConsoleNote(message)
			err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, logfile, message) VALUES (?, ?, ?, ?, ?, ?)`, time.Now().UnixNano(), time.Now().UnixNano(), host, application, logfile, message)
			if err != nil {
				log.Printf("ERROR - can not insert data for logline - %v\n", err)
				http.Error(w, "ERROR", 500); return
			}
			newIDList = append(newIDList, conn.LastInsertRowID())
		}
		file, handler, err := r.FormFile("logfile")
		if err != nil {
			fmt.Printf("No logfile uploaded %v\n", err)
		} else {
			defer file.Close()
			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %s\n", handler.Header["Content-Type"])

			detectContentType := func(out multipart.File) (string, error) {
				buffer := make([]byte, 512)
				_, err := out.Read(buffer)
				if err != nil {
					return "", err
				}
				// Use the net/http package's handy DectectContentType function. Always returns a valid
				// content-type by returning "application/octet-stream" if no others seemed to match.
				contentType := http.DetectContentType(buffer)
				return contentType, nil
			}
			contentType, _ := detectContentType(file); fmt.Printf("DEBUG - detectContentType %s\n",contentType)
			if ! strings.HasPrefix( contentType, "text/") {
				http.Error(w, "Uploaded file is not text tile", http.StatusBadRequest)
			} else {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					logline := scanner.Text()
					logline = FilterPassword(logline, PasswordFilterPtns)
					logline = DecodeJenkinsConsoleNote(logline)
					logfile = Ternary(logfile == "", handler.Filename, logfile).(string)
					err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, logfile, message) VALUES (?, ?, ?, ?, ?, ?)`, time.Now().UnixNano(), time.Now().UnixNano(), host, application, logfile, logline)
					if err != nil {
						log.Printf("ERROR -logline can not insert data for logline - %v\n", err)
						http.Error(w, "ERROR", 500)
					}
					if len(newIDList) < 50{
						newIDList = append(newIDList, conn.LastInsertRowID())
					}
				}
			}
		}
		msg := "<html><body>OK Log saved.<br><ul>"
		for _, id := range(newIDList) {
			url := fmt.Sprintf("https://%s:%d/searchlogbyid/%d", Config.Serverdomain, Config.Port, id)
			l := fmt.Sprintf(`<li><a href="%s">%s</a></li>`, url, url )
			msg = msg + l
		}
		msg = msg + "</ul></body</html>"
		fmt.Fprintf(w, msg)
		return
	}
}

// For debugging purposes only
func DumpPost(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("DEBUG - DUMP\n\n%s\n",requestDump)

	msg, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("DEBUG - Body\n\n%s\n",msg)

	r.Body = ioutil.NopCloser(bytes.NewBuffer(msg))
	fmt.Fprintf(w, "OK")
	return

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
	//webhook - this should only allow request from localhost. the purpose is to allow some dumb app (like jira) to post/get to an url which is https://localhost/wh/<remote_nsre_dns_name>/<command_name>/ - translate the request and use the jwt auth to send request to the remote nsre server to process the command.
	router.HandleFunc("/wh/{proto}/{remote_nsre_host}/{port:[0-9]+}/{CommandName}", SendProcessCommand).Methods("GET", "POST")

	//Probably we should drop the non jwt key section and enforce the use of jwt key all the time
	if Config.JwtKey == "" {
		log.Printf("WARNING WARNING - JWTKEY is not set. Log server will allow anyone to put log in\n")
		router.HandleFunc("/log/{idx_name}/{type_name}/{unique_id}", ProcessLog).Methods("POST")
		router.HandleFunc("/log/{idx_name}/{type_name}", ProcessLog).Methods("POST")
		router.HandleFunc("/log", ProcessLog).Methods("POST")
	} else{
		router.Handle("/log/{idx_name}/{type_name}/{unique_id}", isAuthorized(ProcessLog)).Methods("POST")
		router.Handle("/log/{idx_name}/{type_name}", isAuthorized(ProcessLog)).Methods("POST")
		router.Handle("/log/load", isOauthAuthorized(DoUploadLog)).Methods("POST", "GET")
		router.Handle("/log", isAuthorized(ProcessLog)).Methods("POST")

		//Twilio app
		router.HandleFunc("/twilio/events/{call_id}", ProcessTwilioCallEvent).Methods("POST")
		router.HandleFunc("/twilio/gather/{call_id}", ProcessTwilioGatherEvent).Methods("POST")
		twilioSid := GetConfig("twilio_sid")
		twilioSec := GetConfig("twilio_sec")
		router.Handle("/twilio/{action:(?:call|sms)}", IsBasicAuth(MakeTwilioCall, twilioSid, twilioSec, "Twilio")).Methods("POST")

		//Debugging

		router.Handle("/dump", IsBasicAuth(DumpPost, GetConfig("dump_username", ""), GetConfig("dump_password", ""), "DumpRequest")).Methods("POST", "GET", "PUT")

		//Nagios commands
		router.Handle("/nagios/{command}", isAuthorized(ProcessNagiosCommand)).Methods("POST")
		//Pagerduty event - See the file pagerduty.go for more
		router.Handle("/pagerduty", IsBasicAuth(HandlePagerDutyEvent, GetConfig("pagerduty_username"), GetConfig("pagerduty_password"), "PagerDuty" )).Methods("POST")
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
	CREATE TABLE IF NOT EXISTS log(
		id integer primary key autoincrement,
		timestamp int,
		datelog int,
		host text,
		application text,
		logfile text,
		message text);

	CREATE TABLE IF NOT EXISTS user(id integer primary key autoincrement, username text, email text UNIQUE);
	CREATE UNIQUE INDEX IF NOT EXISTS t_host_idx ON log(timestamp, host, datelog, application);

	CREATE TABLE IF NOT EXISTS appconfig(
		key text,
		val text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS appconfigkeyidx ON appconfig(key);

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
	message := FilterPassword(logData.Message, PasswordFilterPtns)
	message = DecodeJenkinsConsoleNote(message)

	err := conn.Exec(`INSERT INTO log(timestamp, datelog, host, application, logfile, message) VALUES (?, ?, ?, ?, ?, ?)`, logData.Timestamp, logData.Datelog, logData.Host, logData.Application, logData.Logfile, message)
	if err != nil {
		log.Printf("ERROR - can not insert data for logline - %v\n", err)
	}
}

//DatabaseMaintenance - all maintenance routine spawn off from here. Currently just delete the records older than the retention period to keep disk space usage managable. User can always use the sqlite command to export data into csv format and save it somewhere.
func DatabaseMaintenance() {
	conn := GetDBConn()
	defer conn.Close()
	start, _ := ParseTimeRange(Config.LogRetention, "")
	err := conn.Exec(fmt.Sprintf(`DELETE FROM log WHERE timestamp < %d`, start.UnixNano()))
	if err != nil {
		log.Printf("ERROR - can not delete old data - %v\n", err)
	}
}
