package cmd

import (
	"fmt"
	"strings"
	"net/url"
	"crypto/tls"
	"net/http"
	"log"

)

/* The client side of nagios.
These commands using net/http client and make a request to the server run in the nagios server to handle a nagios command. The server side is handled in the file nagios-cmd.go
*/

func DoNagiosDeleteAllComment(Host, Service string) int {
	nagiosNsreBaseURL := GetConfig("nagios_nsre_url", "")
	if nagiosNsreBaseURL == "" { log.Fatalln("ERROR FATAL This feature requires the appconfig key nagios_nsre_url. Please use sqlite3 command to insert a record into the log database. The value of the key is the base url of the nsre instance runs on the nagios server which has the endpoint /nagios/{command} to write nagios command to the command file") }
	validToken, _ := GenerateJWT()

	client := &http.Client{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: Config.IgnoreCertificateCheck}

	command := "del_all_comment"

	formData := url.Values{
		"host": { Host },
		"service": { Service },
	}
	req, _ := http.NewRequest("POST", nagiosNsreBaseURL + "/nagios/" + command, strings.NewReader(formData.Encode()))
	req.Header.Set("Token", validToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR - %v", err)
	}
	defer res.Body.Close()
	return res.StatusCode
}

func DoNagiosACK(Host, Service, User, Comment string) int {
	nagiosNsreBaseURL := GetConfig("nagios_nsre_url", "")
	if nagiosNsreBaseURL == "" { log.Fatalln("ERROR FATAL This feature requires the appconfig key nagios_nsre_url. Please use sqlite3 command to insert a record into the log database. The value of the key is the base url of the nsre instance runs on the nagios server which has the endpoint /nagios/{command} to write nagios command to the command file") }
	validToken, _ := GenerateJWT()

	client := &http.Client{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: Config.IgnoreCertificateCheck}

	command := Ternary(Service == "", "host_ack", "service_ack").(string)

	formData := url.Values{
		"host": { Host },
		"service": { Service },
		"user": {User},
		"comment": {Comment},
	}
	req, _ := http.NewRequest("POST", nagiosNsreBaseURL + "/nagios/" + command, strings.NewReader(formData.Encode()))
	req.Header.Set("Token", validToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR - %v", err)
	}
	defer res.Body.Close()
	return res.StatusCode
}