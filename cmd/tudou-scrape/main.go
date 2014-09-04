package main

import (
	"bitbucket.org/cfstras/tudou/data"
	dl "bitbucket.org/cfstras/tudou/youtube_dl"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/crowdmob/goamz/sqs"
	ct "github.com/daviddengcn/go-colortext"
)

const (
	ErrorNone = iota
	ErrorArgs
	ErrorAwsAuth
	ErrorSource
	ErrorNotImpl
	ErrorGetQueue
	ErrorQueue
	ErrorS3
	ErrorDownload
	ErrorExists
	ErrorFile
)

var stuff struct {
	queueRegionName, bucketRegionName string
	queueName, bucketName             string
	queueRegion, bucketRegion         aws.Region
	sqs                               *sqs.SQS
	auth                              aws.Auth

	items data.ItemSlice
	queue *sqs.Queue
}

func Color(color ct.Color, msg ...interface{}) {
	ct.ChangeColor(color, false, ct.None, false)
	fmt.Print(msg...)
	ct.ResetColor()
}
func Colorln(color ct.Color, msg ...interface{}) {
	ct.ChangeColor(color, false, ct.None, false)
	fmt.Println(msg...)
	ct.ResetColor()
}
func Redln(msg ...interface{}) {
	Colorln(ct.Red, msg...)
}
func Yellow(msg ...interface{}) {
	Color(ct.Yellow, msg...)
}
func Yellowln(msg ...interface{}) {
	Colorln(ct.Yellow, msg...)
}

func main() {
	var doSend, doReceive, help bool
	var jsonSource, tsvSource string

	flag.BoolVar(&help, "help", false, "Show this help")

	flag.BoolVar(&doReceive, "receive", false, "Receive sqs messages and download Videos")
	flag.BoolVar(&doSend, "send", false, "Send sqs messages from source")

	flag.StringVar(&jsonSource, "json", "", "Source: JSON file with array of items")
	flag.StringVar(&tsvSource, "tsv", "", `Source: tsv file with format "video id\tTitle"`)

	flag.StringVar(&stuff.queueName, "queue", "", "SQS Queue Name")
	flag.StringVar(&stuff.bucketName, "bucket", "", "S3 Bucket Name (only for receive)")
	flag.StringVar(&stuff.queueRegionName, "queueRegion", "us-east-1", "AWS region for queue")
	flag.StringVar(&stuff.bucketRegionName, "bucketRegion", "us-west-2", "AWS region for bucket")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if !(doReceive || doSend) || (doReceive && doSend) {
		Redln("Please select receive OR send.")
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if stuff.queueName == "" {
		Redln("Please set queue ID.")
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if doSend && (jsonSource == "" && tsvSource == "") {
		Redln("Please specify JSON or tsv source.")
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if doReceive && stuff.bucketName == "" {
		Redln("Please specify S3 bucket name for receive.")
		flag.Usage()
		os.Exit(ErrorArgs)
	}

	var err error
	stuff.auth, err = aws.EnvAuth()
	if err != nil {
		oldErr := err
		var u *user.User
		u, err = user.Current()
		if err != nil {
			Redln("Could not get User dir and", oldErr)
			os.Exit(ErrorAwsAuth)
		}
		stuff.auth, err = aws.CredentialFileAuth(u.HomeDir+"/.aws/credentials", "default", 0)
		if err != nil {
			Redln("Environment aws auth failed:", oldErr, "- File failed:", err)
			os.Exit(ErrorAwsAuth)
		}
	}

	stuff.queueRegion = getRegion(stuff.queueRegionName)
	stuff.sqs = sqs.New(stuff.auth, stuff.queueRegion)

	stuff.queue, err = stuff.sqs.GetQueue(stuff.queueName)
	if err != nil {
		Redln("GetQueue:", err)
		os.Exit(ErrorGetQueue)
	}
	Yellow("Queue url: ")
	fmt.Println(stuff.queue.Url)

	if doSend {
		if jsonSource != "" {
			err = stuff.items.LoadJSON(jsonSource)
		} else if tsvSource != "" {
			err = stuff.items.LoadTSV(tsvSource)
		}
		if err != nil {
			Err("Error loading JSON/CSV:", err, ErrorSource)
		}
		send()
	}
	if doReceive {
		receive()
	}
	Redln("finished.")
}

func receive() {
	stuff.bucketRegion = getRegion(stuff.bucketRegionName)
	s3client := s3.New(stuff.auth, stuff.bucketRegion)
	/* I haz a */ bucket := s3client.Bucket(stuff.bucketName)

	msgs, err := stuff.queue.ReceiveMessage(1)
	if err != nil {
		Err("Error in ReceiveMessage:", err, ErrorQueue)
	}
	if len(msgs.Messages) < 1 {
		Err("No message received", nil, ErrorQueue)
	}
	msg := &msgs.Messages[0]
	split := strings.Split(msg.Body, "\t")
	if len(split) != 2 {
		Err("Invalid Message in queue: "+msg.Body, nil, ErrorQueue)
	}
	videoId, videoTitle := split[0], split[1]

	// check if file exists already
	if res, err := bucket.List(videoId, "", "", 20); err != nil {
		Err("Error checking S3 for video:", err, ErrorS3)
	} else if len(res.Contents) > 0 {
		gotOne, gotSize, gotKey := false, int64(0), ""
		for _, k := range res.Contents {
			if k.Size > 512*1024 { // 512k min
				gotOne, gotSize, gotKey = true, k.Size, k.Key
				break
			}
		}
		if gotOne {
			_, err = stuff.queue.DeleteMessage(msg)
			if err != nil {
				Err("Error in DeleteMessage:", err, ErrorQueue)
			}
			Err(fmt.Sprint("File already there, size ", gotSize, ", key ",
				gotKey), nil, ErrorExists)
		}
	}

	url := dl.TudouUrl + videoId
	Yellow("Loading ")
	fmt.Println(url)
	file, length, info, infoBytes, err := dl.Load(url)
	if err != nil {
		Err("Error loading video:", err, ErrorDownload)
	}
	path := videoId + " " + videoTitle + "."

	if length == 0 || len(infoBytes) == 0 {
		Err(fmt.Sprint("Error: got invalid lengths. Video size: ", length,
			", Info size: ", len(infoBytes), "; info: ", string(infoBytes)),
			nil, ErrorDownload)
	}

	Yellow("Saving metadata ")
	fmt.Println(path + "json")
	err = bucket.Put(path+"json", infoBytes, "application/json", s3.AuthenticatedRead, s3.Options{})
	if err != nil {
		Err("Error putting metadata:", err, ErrorQueue)
	}

	Yellowln("Saving video ...")
	err = bucket.PutReader(path+info.Extension, file, length, "video/x-flv", s3.AuthenticatedRead, s3.Options{})
	if err != nil {
		Err("Error PUTting video:", err, ErrorS3)
	}
	Yellowln("Removing tempfiles...")
	err = file.Close()
	if err != nil {
		Err("Error closing file:", err, ErrorFile)
	}
	err = os.Remove(file.Name())
	if err != nil {
		Err("Error deleting file:", err, ErrorFile)
	}

	_, err = stuff.queue.DeleteMessage(msg)
	if err != nil {
		Err("Error in DeleteMessage:", err, ErrorQueue)
	}
}

func send() {
	for _, item := range stuff.items {
		msg := item.Code + "\t" + item.Title
		resp, err := stuff.queue.SendMessage(msg)
		if err != nil {
			Redln("SendMessage:", err)
			os.Exit(ErrorQueue)
		}
		Yellow("Sent ", msg)
		fmt.Println(" id:", resp.Id, "md5:", resp.MD5)
	}
}

func Err(msg string, err error, exitCode int) {
	Redln(msg, err)
	os.Exit(exitCode)
}

func getRegion(name string) (r aws.Region) {
	var ok bool
	r, ok = aws.Regions[name]
	if !ok {
		Redln("Region", name, "not supported. Available are:")
		for k := range aws.Regions {
			fmt.Print(k, ", ")
		}
		Redln()
		os.Exit(ErrorArgs)
	}
	return
}
