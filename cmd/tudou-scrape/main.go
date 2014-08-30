package main

import (
	"bitbucket.org/cfstras/tudou/data"
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/sqs"
)

const (
	ErrorNone = iota
	ErrorArgs
	ErrorAwsAuth
	ErrorSource
	ErrorNotImpl
	ErrorGetQueue
	ErrorSendMessage
)

var stuff struct {
	queueId, region string
	sqs             *sqs.SQS
	auth            aws.Auth

	items data.ItemSlice
}

func main() {
	var doSend, doReceive, help bool
	var jsonSource, tsvSource string

	flag.BoolVar(&help, "help", false, "Show this help")

	flag.BoolVar(&doReceive, "receive", false, "Receive sqs messages and download Videos")
	flag.BoolVar(&doSend, "send", false, "Send sqs messages from source")

	flag.StringVar(&jsonSource, "json", "", "Source: JSON file with array of items")
	flag.StringVar(&tsvSource, "tsv", "", `Source: tsv file with format "video id\tTitle"`)

	flag.StringVar(&stuff.queueId, "queue", "", "SQS queue ID")
	flag.StringVar(&stuff.region, "region", "us-east-1", "AWS Region")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if !(doReceive || doSend) || (doReceive && doSend) {
		fmt.Println("Please select receive OR send")
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if stuff.queueId == "" {
		fmt.Println("Please set queue ID")
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if doSend && (jsonSource == "" && tsvSource == "") {
		fmt.Println("Please specify JSON or tsv source.")
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
			fmt.Println("Could not get User dir and", oldErr)
			os.Exit(ErrorAwsAuth)
		}
		stuff.auth, err = aws.CredentialFileAuth(u.HomeDir+"/.aws/credentials", "default", 0)
		if err != nil {
			fmt.Println("Environment aws auth failed:", oldErr, "- File failed:", err)
			os.Exit(ErrorAwsAuth)
		}
	}

	re, ok := aws.Regions[stuff.region]
	if !ok {
		fmt.Println("Region", stuff.region, "not supported. Available are:")
		for k := range aws.Regions {
			fmt.Print(k, ", ")
		}
		fmt.Println()
		os.Exit(ErrorArgs)
	}
	stuff.sqs = sqs.New(stuff.auth, re)

	if doSend {
		if jsonSource != "" {
			err = stuff.items.LoadJSON(jsonSource)
		} else if tsvSource != "" {
			err = stuff.items.LoadTSV(tsvSource)
		}
		if err != nil {
			fmt.Println("Loading JSON/CSV:", err)
			os.Exit(ErrorSource)
		}
		send()
	}
	if doReceive {
		receive()
	}
	fmt.Println("finished.")
}

func receive() {
	fmt.Println("Error: not implemented")
	os.Exit(ErrorNotImpl)
}

func send() {
	queue, err := stuff.sqs.GetQueue(stuff.queueId)
	if err != nil {
		fmt.Println("GetQueue:", err)
		os.Exit(ErrorGetQueue)
	}
	for _, item := range stuff.items {
		msg := item.Code + "\t" + item.Title
		resp, err := queue.SendMessage(msg)
		if err != nil {
			fmt.Println("SendMessage:", err)
			os.Exit(ErrorSendMessage)
		}
		fmt.Println("Sent", msg, "id:", resp.Id, "md5:", resp.MD5)
	}
}
