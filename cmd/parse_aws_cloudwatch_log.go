package cmd

import (
	"time"
	"regexp"
	"log"
	"bufio"
	"os"
	"github.com/json-iterator/go"
)

//AWSLogEvent
type AWSLogEvent struct {
	TimeStamp int64
	Message string
	IngestionTime int64
}

//AWSLog
type AWSLog struct {
	Events []AWSLogEvent
	NextForwardToken string
	NextBackwardToken string
}

//ParseAWSCloudWatchLogEvent - Input from os.stdin as a json string with the format describe in aws cloudwatch logevent. It is the output of command such as aws logs get-log-events --log-group-name /aws/ecs/int --log-stream-name errcd-wa-int/errcd-wa-task/1bd0169e-2013-4964-96fa-8c18819ffa62 --profile errcd_wa --region ap-southeast-2
//We parse it and ship it to out log server.
func ParseAWSCloudWatchLogEvent(appNameStr string) {
	reader := bufio.NewReader(os.Stdin)
	//1048576 Maxium aws cloudwatch event size
	buff := make([]byte, 1048576)
	reader.Read(buff)

	awsLog := AWSLog{}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if e := json.Unmarshal(buff, &awsLog); e != nil {
		log.Fatalf("ERROR parsing awslog event data - %v\n", e)
	}

	for _, data := range(awsLog.Events) {
		hostStr, _ := os.Hostname()
		timeHarvest, timeParsed  := time.Now(), MsToTime(data.TimeStamp)
		logFile, msgStr, passPtnStr :=  "stdin", data.Message, Config.PasswordFilterPattern
		passPtn := regexp.MustCompile(passPtnStr)
		SendLine(timeHarvest, timeParsed, hostStr, appNameStr, logFile, msgStr, passPtn)
	}

}