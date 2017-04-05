package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var version string
var printVersion bool

func main() {
	flagBucket := flag.String("bucket", "", "bucket name")
	flagPrefix := flag.String("prefix", "", "prefix in the bucket")
	flagMinWarn := flag.String("min-warn", "-1", "minimum size for warning, in bytes or with k/M/G suffix")
	flagMaxWarn := flag.String("max-warn", "-1", "maximum size for warning, in bytes or with k/M/G suffix")
	flagMinCrit := flag.String("min-crit", "-1", "minimum size for critical, in bytes or with k/M/G suffix")
	flagMaxCrit := flag.String("max-crit", "-1", "maximum size for critical, in bytes or with k/M/G suffix")
	flag.BoolVar(&printVersion, "V", false, "print version and exit")
	flag.Parse()

	if printVersion {
		fmt.Printf("Version: %s\n", version)
		os.Exit(int(ReturnCode(OK)))
	}

	if *flagBucket == "" {
		fmt.Println("-bucket is required")
		execError()
	}

	minBytesWarn, err := calculate(*flagMinWarn)
	if err != nil {
		execError()
	}
	maxBytesWarn, err := calculate(*flagMaxWarn)
	if err != nil {
		execError()
	}
	minBytesCrit, err := calculate(*flagMinCrit)
	if err != nil {
		execError()
	}
	maxBytesCrit, err := calculate(*flagMaxCrit)
	if err != nil {
		execError()
	}

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
	err = svc.ListObjectsPages(params, func(page *s3.ListObjectsOutput, lastPage bool) bool {
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

func execError() {
	flag.Usage()
	os.Exit(int(ReturnCode(CRITICAL)))
}

func calculate(input string) (int64, error) {
	if input == "" || input == "-1" {
		return int64(-1), nil
	}
	r, err := regexp.Compile("([0-9]+)([kMG]?)")
	if err != nil {
		return int64(-1), errors.New("Error compiling regex")
	}
	var result int
	var numeric int
	matches := r.FindStringSubmatch(input)

	if len(matches[2]) > 0 {
		numeric, err = strconv.Atoi(matches[1])
		if result < 0 || err != nil {
			return int64(-1), errors.New("invalid result")
		}
		unit := matches[2]
		if unit == "k" {
			result = numeric * 1024
		} else if unit == "M" {
			result = numeric * 1024 * 1024
		} else if unit == "G" {
			result = numeric * 1024 * 1024 * 1024
		} else {
			return int64(-1), errors.New("invalid result")
		}
	} else {
		result, err = strconv.Atoi(input)
		if result < 0 || err != nil {
			return int64(-1), errors.New("invalid result")
		}
	}

	return int64(result), nil
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
