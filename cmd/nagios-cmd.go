package cmd

import (
	"os"
	"io/ioutil"
	"time"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)
/* This is to implement nagios command such as host/service ACKNOWLEDGE or something else.
It run the same host as the nagios server thus it can acccess the nagios cmd file.
It takes the request from http(s) endpoint via a remote nsre client (or any http client that has the jwt auth token)

My intention is that after the Twilio call updated, user press key eg. 4- it sends back to the twilio proxy server. That server will based on keypad value to send a nagios command to this instance to control nagios state.

nagios runs within private network for security reason thus the nsre acting as twilio proxy should be in the same private network.

The bigger picture design.

In a company we would have on external accessible nsre act as log collection server, twilio proxy server.
And for each server in internal private network we have at least one nsre runs for log shipper, and through nrse features we can run nagsio commands checks (predefined list).
On the nagios server we can run nagios command

all of it uses this program and it is small/fast/multipurposes but loosely coupled for maintenance.

See https://assets.nagios.com/downloads/nagioscore/docs/externalcmds/index.php for nagios doco

*/

//ProcessNagiosCommand - entry point taken for /nagios/{command}. We will route each command to the real handler
func ProcessNagiosCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commandName := vars["command"]
	switch commandName {
	case "service_ack":
		HandleNagiosServiceACK(&w, r)
	case "host_ack":
		HandleNagiosHostACK(&w, r)
	case "del_all_comment":
		HandleNagiosDeleteAllComment(&w, r)
	}
}

func HandleNagiosServiceACK(w *http.ResponseWriter, r *http.Request) {
	nagiosCmdFile := GetConfigSave("nagios_cmd_file", "/var/spool/nagios/cmd/nagios.cmd")
	host := r.FormValue("host")
	serviceDecs := r.FormValue("service")
	user := r.FormValue("user")
	comment := r.FormValue("comment")
	comment = Ternary(comment == "", "Acknowledgement From Twilio", comment).(string)
	data := fmt.Sprintf("[%d] ACKNOWLEDGE_SVC_PROBLEM;%s;%s;2;1;1;%s;%s\n", time.Now().Unix(), host, serviceDecs, user, comment)
	fmt.Printf("DEBUG data going to send to nagios '%s'\n", data)
    if err := ioutil.WriteFile(nagiosCmdFile, []byte(data), 0644); err != nil {
		fmt.Printf("ERROR writting to nagios cmd file - %v\n", err)
		http.Error(*w, "ERROR", 500); return
	}
	fmt.Fprintf(*w, "OK"); return
}

func HandleNagiosHostACK(w *http.ResponseWriter, r *http.Request) {
	nagiosCmdFile := GetConfigSave("nagios_cmd_file", "/var/spool/nagios/cmd/nagios.cmd")
	host := r.FormValue("host")
	user := r.FormValue("user")
	data := fmt.Sprintf("[%d] ACKNOWLEDGE_HOST_PROBLEM;%s;2;1;1;%s;Acknowledgement From Twilio\n", time.Now().Unix(), host, user)
    if err := ioutil.WriteFile(nagiosCmdFile, []byte(data), 0644); err != nil {
		fmt.Printf("ERROR writting to nagios cmd file - %v\n", err)
		http.Error(*w, "ERROR", 500); return
	}
	fmt.Fprintf(*w, "OK"); return
}

func sendError(w *http.ResponseWriter, msg string) { fmt.Printf(msg); http.Error(*w, "ERROR", 500) }

func HandleNagiosDeleteAllComment(w *http.ResponseWriter, r *http.Request) {
	nagiosCmdFile := GetConfigSave("nagios_cmd_file", "/var/spool/nagios/cmd/nagios.cmd")
	host := r.FormValue("host")
	serviceDecs := r.FormValue("service")
	var data string
	if serviceDecs == "" {
		data = fmt.Sprintf("[%d] DEL_ALL_HOST_COMMENTS;%s\n", time.Now().Unix(), host)
	} else {
		data = fmt.Sprintf("[%d] DEL_ALL_SVC_COMMENTS;%s;%s\n", time.Now().Unix(), host, serviceDecs)
	}

	fi, e := os.Stat(nagiosCmdFile)

	if e != nil { sendError(w, "ERROR unexpected nagios cmd file - " + e.Error() + "\n" ); return }

	if fi.Mode()&os.ModeNamedPipe != 0 {
		if err := ioutil.WriteFile(nagiosCmdFile, []byte(data), 0644); err != nil {
			sendError(w, "ERROR writting to nagios cmd file - " + err.Error() + "\n" ); return
		}
		fmt.Fprintf(*w, "OK"); return
	} else { sendError(w, "ERROR unexpected nagios cmd file. It is not a pipe \n" ); return	}
}