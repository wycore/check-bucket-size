package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	flagBucket := flag.String("bucket", "my-bucket", "bucket name")
	flagPrefix := flag.String("prefix", "", "prefix in the bucket")
	flagMinBytesWarn := flag.Int("min-bytes-warn", -1, "min-bytes warn")
	flagMaxBytesWarn := flag.Int("max-bytes-warn", -1, "max-bytes warn")
	flagMinBytesCrit := flag.Int("min-bytes-crit", -1, "min-bytes crit")
	flagMaxBytesCrit := flag.Int("max-bytes-crit", -1, "max-bytes crit")
	flag.Parse()

	minBytesWarn := int64(*flagMinBytesWarn)
	maxBytesWarn := int64(*flagMaxBytesWarn)
	minBytesCrit := int64(*flagMinBytesCrit)
	maxBytesCrit := int64(*flagMaxBytesCrit)

	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc := s3.New(sess, &aws.Config{
		Region: aws.String("eu-west-1"),
	})

	params := &s3.ListObjectsInput{
		Bucket: aws.String(*flagBucket),
		Prefix: aws.String(*flagPrefix),
	}

	size := int64(0)

	pageNum := 0
	err := svc.ListObjectsPages(params, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, item := range page.Contents {
			size += *item.Size
		}
		pageNum++
		// fmt.Println(len(page.Contents))
		return true
	})
	if err != nil {
		writeCheckOutput(ReturnCode(UNKNOWN), fmt.Sprintf("Unable to list contents of bucket s3://%s", *flagBucket), "")
	}

	if minBytesWarn > int64(-1) && size < minBytesWarn {
		writeCheckOutput(ReturnCode(WARNING), fmt.Sprintf("Contents too small: s3://%s/%s", *flagBucket, *flagPrefix), "")
	}
	if maxBytesWarn > int64(-1) && size > maxBytesWarn {
		writeCheckOutput(ReturnCode(WARNING), fmt.Sprintf("Contents too big: s3://%s/%s", *flagBucket, *flagPrefix), "")
	}
	if minBytesCrit > int64(-1) && size < minBytesCrit {
		writeCheckOutput(ReturnCode(CRITICAL), fmt.Sprintf("Contents too small: s3://%s/%s", *flagBucket, *flagPrefix), "")
	}
	if maxBytesCrit > int64(-1) && size > maxBytesCrit {
		writeCheckOutput(ReturnCode(CRITICAL), fmt.Sprintf("Contents too big: s3://%s/%s", *flagBucket, *flagPrefix), "")
	}

	writeCheckOutput(ReturnCode(OK), "OK", "")
}

func writeCheckOutput(code ReturnCode, message string, additional string) {
	var prefix = "Check "
	var result = ""
	switch code {
	case OK:
		result = "OK"
	case WARNING:
		result = "WARNING"
	case CRITICAL:
		result = "CRITICAL"
	default:
		result = "UNKNOWN"
	}
	fmt.Printf("%s%s: %s\n", prefix, result, message)
	if len(additional) > 0 {
		fmt.Printf("%s\n", additional)
	}
	os.Exit(int(code))
}
