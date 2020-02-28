package cmd

import (
	"fmt"
	"strings"
	"io/ioutil"
	"net/http"
)
/* Process pager duty generic webhook v2
The reason is Pagerduty Nagios integeration is buggy in the way back from pager duty to nagios, I never be able to make it work and pager duty staff just ignore about it. It never post anything!

But their generic webhook v2 seems to post event fine (use the /dump from this app to confirm that). So we can use this to send Nagios ACK - and this is the goal of this file.

The endpoint would be /pagerduty

To use this feature you need to insert into the appconfig table the key pagerduty_user, pagerduty_password with a random string. Then use it as pagerduty webhook url like https://user:password@your_nsre_dns/pagerduty
*/

func HandlePagerDutyEvent(w http.ResponseWriter, r *http.Request) {
	bodyDataByte, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("DEBUG HandlePagerDutyEvent called with body data\n\n%s\n", string(bodyDataByte))
	event := json.Get(bodyDataByte, "messages", 0, "event").ToString()
	switch event {
	case "incident.acknowledge":
		ProcessACKEvent(&w, bodyDataByte)
	case "incident.resolve":
		ProcessResolvedEvent(&w, bodyDataByte)
	default:
		fmt.Printf("HandlePagerDutyEvent EVENT UNHANDLED - event '%s'\n", event) ; fmt.Fprintf(w, "EVENT UNHANDLED"); return
	}
	fmt.Fprintf(w, "OK"); return
}

func ProcessResolvedEvent(w *http.ResponseWriter, bodyDataByte []byte) {
	fmt.Fprintf(*w, "TODO"); return
}

func ProcessACKEvent(w *http.ResponseWriter, bodyDataByte []byte) {
	alertKey := json.Get(bodyDataByte, "messages", 0, "incident", "alerts", 0, "alert_key").ToString()
	var event_source, host_name, service_desc string
	// sample text "event_source=service;host_name=xvt-aws-ansible;service_desc=check_xvt_services"
	for _, item := range (strings.Split(alertKey, ";")) {
		itemEqual := strings.Split(item, "=")
		if len(itemEqual) != 2 { fmt.Printf("ERROR HandlePagerDutyEvent - Input line '%s'. Event source: %s\n", alertKey, event_source); fmt.Fprintf(*w, "UNEXPECTED DATA"); return }
		switch itemEqual[0] {
		case "event_source":
			event_source = itemEqual[1]
		case "host_name":
			host_name = itemEqual[1]
		case "service_desc":
			service_desc = itemEqual[1]
		}
	}
	if host_name == "" { fmt.Printf("ERROR HandlePagerDutyEvent - Input line '%s'\n", alertKey); fmt.Fprintf(*w, "UNEXPECTED DATA"); return }
	comment := json.Get(bodyDataByte, "messages", 0, "log_entries", 0, "summary").ToString()
	fmt.Printf("DEBUG going to call DoNagiosACK with data '%s' - '%s' - '%s'\n", host_name, service_desc, comment)
	StatusCode := DoNagiosACK(host_name, service_desc, "PagerDutyPostBack", comment)
	if StatusCode != 200 {
		fmt.Fprintf(*w, "ERROR when talking to nagios cmd."); return
	}
}