package cmd

import (
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"net/url"
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
	SendAWSLogEvents(awsLog.Events, appNameStr, 0, nil)
}

func StartAllAWSCloudwatchLogPolling(wg *sync.WaitGroup, c chan struct{}) {
	for _, cfg := range(Config.AWSLogs) {
		wg.Add(1)
		log.Printf("Start parsing aws log config %v\n", cfg)
		go StartAWSCloudwatchLogPolling(&cfg, wg)
	}
	<-c
	log.Printf("Signal captured. Do cleaning up\n")
	wg.Done()
}

//StartAWSCloudwatchLogOnePrefix -
func StartAWSCloudwatchLogOnePrefix(cfg *AWSLogConfig, cl *cloudwatchlogs.CloudWatchLogs, filterEvtInput *cloudwatchlogs.FilterLogEventsInput, sleepDuration time.Duration) {
	var lastEndTime int64

	var conn *sqlite3.Conn

	u, err := url.Parse(Config.Serverurl)
	if err != nil {
		log.Fatal(err)
	}

	if u.Hostname() == Config.Serverdomain {
		conn = GetDBConn()
		defer conn.Close()
	}

	for {
		if lastEndTime != 0 {
			filterEvtInput.SetStartTime(lastEndTime)
			now := time.Now().UnixNano() / NanosPerMillisecond
			filterEvtInput.SetEndTime(now)
		}
		out, e := cl.FilterLogEvents(filterEvtInput)
		if e != nil {
			log.Printf("ERROR can not FilterLogEvent. Maybe api throttling. Sleep 15 minutes - %v\n", e)
			time.Sleep(15 * time.Minute)
		}
		events := out.Events
		lastEndTime = SendAWSLogEvents(events, cfg.LoggroupName, lastEndTime, conn)

		log.Printf("Sleep %v\n", sleepDuration)
		time.Sleep(sleepDuration)
	}
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

	for _, streamPrefix := range(cfg.StreamPrefix) {
		filterEvtInput := cloudwatchlogs.FilterLogEventsInput{
			StartTime: &startInMs,
			EndTime: &endInMs,
			LogGroupName: &cfg.LoggroupName,
			LogStreamNamePrefix: &streamPrefix,
			FilterPattern: &cfg.FilterPtn,
		}
		go StartAWSCloudwatchLogOnePrefix(cfg, clog, &filterEvtInput, sleepDuration)
	}
	wg.Done()
}