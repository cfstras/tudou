package main

import (
	. "bitbucket.org/cfstras/tudou/color"
	"flag"
	"fmt"
	"os"
	"os/user"
	"regexp"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

var stuff struct {
	bucketName, bucketRegionName string
	bucketRegion                 aws.Region
	auth                         aws.Auth
	s3client                     *s3.S3
	bucket                       *s3.Bucket
}

var split = regexp.MustCompile(`[ \.]`)

const (
	ErrorNone int = iota
	ErrorArgs
	ErrorAwsAuth
	ErrorBucket
	ErrorDelete
)

func main() {
	var help bool
	flag.StringVar(&stuff.bucketRegionName, "bucketRegion", "us-west-2", "AWS region for bucket")
	flag.StringVar(&stuff.bucketName, "bucket", "", "S3 Bucket Name (only for receive)")
	flag.BoolVar(&help, "help", false, "Show this help")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(ErrorArgs)
	}
	if stuff.bucketName == "" {
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
	stuff.bucketRegion = getRegion(stuff.bucketRegionName)
	stuff.s3client = s3.New(stuff.auth, stuff.bucketRegion)
	stuff.bucket = stuff.s3client.Bucket(stuff.bucketName)

	marker := ""

	for {
		res, err := stuff.bucket.List("", "", marker, 1000)
		if err != nil {
			die("List from "+marker, err, ErrorBucket)
		}

		for _, f := range res.Contents {
			parts := split.Split(f.Key, -1)
			marker = f.Key
			if len(parts) > 2 {
				n := parts[0] + "." + parts[len(parts)-1]
				Yellow("Rename ")
				fmt.Println(f.Key, "to", n)
				_, err = stuff.bucket.PutCopy(n, s3.AuthenticatedRead,
					s3.CopyOptions{}, stuff.bucketName+"/"+f.Key)
				if err != nil {
					Redln("Rename Error:", err)
					continue
				}
				deleted := false
				for tries := 1; !deleted && tries < 5; tries++ {
					err = stuff.bucket.Del(f.Key)
					if err != nil {
						Yellowln("Delete try", tries, "failed")
					} else {
						deleted = true
					}
				}
				if !deleted {
					die("Error deleting "+f.Key, nil, ErrorDelete)
				}
			} else {
				Yellow("will not rename ")
				fmt.Println(f.Key)
			}
		}
		if len(res.Contents) == 0 {
			break
		}
	}
	Yellowln("finished.")
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

func die(msg string, err error, exitCode int) {
	Redln(msg, err)

	os.Exit(exitCode)
}
