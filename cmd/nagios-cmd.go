package cmd

import (
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
	}
}

func HandleNagiosServiceACK(w *http.ResponseWriter, r *http.Request) {
	nagiosCmdFile := GetConfigSave("nagios_cmd_file", "/var/spool/nagios/cmd/nagios.cmd")
	host := r.FormValue("host")
	serviceDecs := r.FormValue("service")
	user := r.FormValue("user")
	data := fmt.Sprintf("[%d] ACKNOWLEDGE_SVC_PROBLEM;%s;%s;2;1;1;%s;Acknowledgement From Twilio\n", time.Now().Unix(), host, serviceDecs, user)
	fmt.Printf("DEBUG data going to send to nagios '%s'\n", data)
    if err := ioutil.WriteFile(nagiosCmdFile, []byte(data), 0644); err != nil {
		fmt.Printf("ERROR writting to nagios cmd file - %v\n", err)
		http.Error(*w, "ERROR", 500); return
	}
	fmt.Fprintf(*w, "OK"); return
}

func HandleNagiosHostACK(w *http.ResponseWriter, r *http.Request) {
	nagiosCmdFile := GetConfigSave("nagios_cmd_file", "/var/spool/nagios/cmd/nagios.cmd")
	r.ParseForm()
	host := r.FormValue("host")
	user := r.FormValue("user")
	data := fmt.Sprintf("[%d] ACKNOWLEDGE_HOST_PROBLEM;%s;2;1;1;%s;Acknowledgement From Twilio\n", time.Now().Unix(), host, user)
    if err := ioutil.WriteFile(nagiosCmdFile, []byte(data), 0644); err != nil {
		fmt.Printf("ERROR writting to nagios cmd file - %v\n", err)
		http.Error(*w, "ERROR", 500); return
	}
	fmt.Fprintf(*w, "OK"); return
}