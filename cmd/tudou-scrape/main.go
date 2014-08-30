package main

import (
	"bitbucket.org/cfstras/tudou/data"
	"flag"
	"fmt"
	"os"
	"os/user"

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
	ErrorSendMessage
	ErrorS3
)

var stuff struct {
	regionName            string
	queueName, bucketName string
	region                aws.Region
	sqs                   *sqs.SQS
	auth                  aws.Auth

	items data.ItemSlice
	queue *sqs.Queue
}

func Color(color ct.Color, msg ...interface{}) {
	ct.ChangeColor(color, false, ct.None, false)
	fmt.Print(msg...)
	ct.ResetColor()
}
func Colorln(color ct.Color, msg ...interface{}) {
	msg2 := make([]interface{}, 0, len(msg)*2-1)
	for i, s := range msg {
		if i != 0 {
			msg2 = append(msg2, " ")
		}
		msg2 = append(msg2, s)
	}
	Color(color, msg2...)
	fmt.Println()
}
func Redln(msg ...interface{}) {
	Colorln(ct.Red, msg...)
}
func Yellow(msg ...interface{}) {
	Color(ct.Yellow, msg...)
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
	flag.StringVar(&stuff.regionName, "region", "us-east-1", "AWS Region")

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

	var ok bool
	stuff.region, ok = aws.Regions[stuff.regionName]
	if !ok {
		Redln("Region", stuff.regionName, "not supported. Available are:")
		for k := range aws.Regions {
			fmt.Print(k, ", ")
		}
		Redln()
		os.Exit(ErrorArgs)
	}
	stuff.sqs = sqs.New(stuff.auth, stuff.region)

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
			Redln("Loading JSON/CSV:", err)
			os.Exit(ErrorSource)
		}
		send()
	}
	if doReceive {
		receive()
	}
	Redln("finished.")
}

func receive() {
	s3client := s3.New(stuff.auth, stuff.region)
	/* I haz a */ bucket := s3client.Bucket(stuff.bucketName)
	res, err := bucket.List("", "", "", 100)
	if err != nil {
		Redln("S3 list:", err)
		os.Exit(ErrorS3)
	}
	Redln("res:", res.Contents)

	//TODO
	Redln("Error: not implemented")
	os.Exit(ErrorNotImpl)
}

func send() {
	for _, item := range stuff.items {
		msg := item.Code + "\t" + item.Title
		resp, err := stuff.queue.SendMessage(msg)
		if err != nil {
			Redln("SendMessage:", err)
			os.Exit(ErrorSendMessage)
		}
		Yellow("Sent", msg)
		fmt.Println("id:", resp.Id, "md5:", resp.MD5)
	}
}
