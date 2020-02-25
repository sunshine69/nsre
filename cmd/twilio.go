package cmd

import (
	"net/http"
	"fmt"
)

// Simple twilio for nagios call
// This to allow place call/sms with text to twillio. Twillio api call is a fire off thing, we need to query the state and handle it properly.

//This does not intend to be full featured. Instead it tries to keep simple and just to be used for nagios notification only

//This app will create a listener /twilio/status_callback to take the status call back from Twillio
// /twilio/call|sms - Make a call or sms
// It would use the existing LogData database to log the call state queue and re-try if failed state occured

func MakeTwilioCall(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	Body := r.FormValue("Body")
	From := r.FormValue("From")
	To := r.FormValue("To")
	fmt.Printf("DEBUG Body: %s - From: %s - To: %s\n", Body, From, To)
}