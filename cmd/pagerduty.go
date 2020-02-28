package cmd

import (
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/gorilla/mux"
	"net/http"
)
/* Process pager duty generic webhook v2
The reason is Pagerduty Nagios integeration is buggy in the way back from pager duty to nagios, I never be able to make it work and pager duty staff just ignore about it. It never post anything!

But their generic webhook v2 seems to post event fine (use the /dump from this app to confirm that). So we can use this to sen Nagios ACK - and this is the goal of this file.

The endpoint would be /pagerduty/{pagerduty_shared_key}
share_key is poorman authentication as the limitation of webhook - they dont provide anything else.

To use need to insert into the appconfig table the key pagerduty_shared_key with a random string. Then use it as pagerduty webhook url like https://your_nsre_dns/pagerduty/<shared_key>
*/

func HandlePagerDutyEvent(w http.ResponseWriter, r *http.Request) {
	sharedKey := GetConfig("pagerduty_shared_key")
	if sharedKey != mux.Vars(r)["pagerduty_shared_key"] {
		http.Error(w, "ERROR", 403)
		return
	}
	bodyDataByte, _ := ioutil.ReadAll(r.Body)
	event := json.Get(bodyDataByte, "messages", 0, "event").ToString()
	if event != "incident.acknowledge" { http.Error(w, "EVENT UNHANDLED",501); return	 }
	alertKey := json.Get(bodyDataByte, "messages", 0, "incident", "alerts", 0, "alert_key").ToString()
	var event_source, host_name, service_desc string
	// sample text "event_source=service;host_name=xvt-aws-ansible;service_desc=check_xvt_services"
	for _, item := range (strings.Split(alertKey, ";")) {
		itemEqual := strings.Split(item, "=")
		if len(itemEqual) != 2 { fmt.Printf("ERROR HandlePagerDutyEvent - Input line '%s'. Event source: %s\n", alertKey, event_source); http.Error(w, "UNEXPECTED DATA", 501); return }
		switch itemEqual[0] {
		case "event_source":
			event_source = itemEqual[1]
		case "host_name":
			host_name = itemEqual[1]
		case "service_desc":
			service_desc = itemEqual[1]
		}
	}
	if host_name == "" { fmt.Printf("ERROR HandlePagerDutyEvent - Input line '%s'\n", alertKey); http.Error(w, "UNEXPECTED DATA", 501); return }
	comment := json.Get(bodyDataByte, "messages", 0, "log_entries", 0, "summary").ToString()
	StatusCode := DoNagiosACK(host_name, service_desc, "PagerDutyPostBack", comment)
	if StatusCode != 200 {
		http.Error(w, "ERROR when talking to nagios cmd.", 500); return
	}
	fmt.Fprintf(w, "OK"); return
}