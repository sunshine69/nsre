package cmd

import (
	"os/signal"

	"sync"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"bufio"
	"os"
	"github.com/json-iterator/go"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

//ParseAWSCloudWatchLogEvent - Input from os.stdin as a json string with the format describe in aws cloudwatch logevent. It is the output of command such as aws logs get-log-events --log-group-name /aws/ecs/int --log-stream-name errcd-wa-int/errcd-wa-task/1bd0169e-2013-4964-96fa-8c18819ffa62 --profile errcd_wa --region ap-southeast-2
//We parse it and ship it to out log server.
func ParseAWSCloudWatchLogEvent(appNameStr string) {
	reader := bufio.NewReader(os.Stdin)
	//1048576 Maxium aws cloudwatch event size
	buff := make([]byte, 1048576)
	reader.Read(buff)

	awsLog := cloudwatchlogs.FilterLogEventsOutput{}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if e := json.Unmarshal(buff, &awsLog); e != nil {
		log.Fatalf("ERROR parsing awslog event data - %v\n", e)
	}

	SendAWSLogEvents(awsLog.Events, appNameStr, 0)
}

func StartAllAWSCloudwatchLogPolling(wg *sync.WaitGroup, c chan struct{}) {
	for _, cfg := range(Config.AWSLogs) {
		wg.Add(1)
		log.Printf("Start parsing aws log config %v\n", cfg)
		go StartAWSCloudwatchLogPolling(&cfg, wg)
	}
	<-c
	log.Printf("Signal captured. Do cleaning up\n")
	signal.Reset()
	wg.Done()
}

//StartAWSCloudwatchLogPolling -
func StartAWSCloudwatchLogPolling(cfg *AWSLogConfig, wg *sync.WaitGroup) {
	region := cfg.Region
	if region == "" { region = "ap-southeast-2" }
	ses, e := session.NewSessionWithOptions(session.Options{
		Profile: cfg.Profile,
		// Provide SDK Config options, such as Region.
		Config: aws.Config {
			Region: &region,
			// Credentials: cred,
		},
	})
	if e != nil {
		log.Fatalf("ERROR can not create session - %v\n", e)
	}

	clog := cloudwatchlogs.New(ses)
	start, end := ParseTimeRange(cfg.Period, CurrentZone)
	startInMs := start.UnixNano() / NanosPerMillisecond
	endInMs := end.UnixNano() / NanosPerMillisecond

	sleepDuration := end.Sub(start)

	var lastEndTime int64

	for {
		fInput := cloudwatchlogs.FilterLogEventsInput{
			StartTime: &startInMs,
			EndTime: &endInMs,
			LogGroupName: &cfg.LoggroupName,
			LogStreamNamePrefix: &cfg.StreamPrefix,
			FilterPattern: &cfg.FilterPtn,
		}

		// log.Printf("last endtime: %d\n",lastEndTime)
		if lastEndTime != 0 {
			fInput.SetStartTime(lastEndTime)
			now := time.Now().UnixNano() / NanosPerMillisecond
			fInput.SetEndTime(now)
		}

		out, e := clog.FilterLogEvents(&fInput)
		if e != nil {
			log.Fatalf("ERROR can not FilterLogEvent - %v\n", e)
		}

		events := out.Events
		lastEndTime = SendAWSLogEvents(events, cfg.LoggroupName, lastEndTime)
		log.Printf("Sleep %v\n", sleepDuration)
		time.Sleep(sleepDuration)
	}
	wg.Done()
}