package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"google.golang.org/api/iterator"
)

var version string
var printVersion bool

// Amazon S3 Provider
const PROVIDER_S3 = "s3"
// Google Cloud Storage Provider
const PROVIDER_GS = "gs"
// Protocol for the bucket URLs - Amazon S3
const PROTO_S3 = "s3"
// Protocol for the bucket URLs - Google Cloud Storage
const PROTO_GS = "gs"

func main() {
	flagBucket := flag.String("bucket", "", "bucket name")
	flagPrefix := flag.String("prefix", "", "prefix in the bucket")
	flagProvider := flag.String("provider", "", "'s3' for Amazon S3 or 'gs' for Google Cloud Storage")
	flagMinWarn := flag.String("min-warn", "-1", "minimum size for warning, in bytes or with k/M/G suffix")
	flagMaxWarn := flag.String("max-warn", "-1", "maximum size for warning, in bytes or with k/M/G suffix")
	flagMinCrit := flag.String("min-crit", "-1", "minimum size for critical, in bytes or with k/M/G suffix")
	flagMaxCrit := flag.String("max-crit", "-1", "maximum size for critical, in bytes or with k/M/G suffix")
	flagDebug := flag.Bool("debug", false, "Show debug output")
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

	if *flagProvider != PROVIDER_S3 && *flagProvider != PROVIDER_GS {
		fmt.Println("-provider is required")
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

	size := int64(0)
	protocol := ""

	if *flagProvider == PROVIDER_S3 {
		protocol = PROTO_S3
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

		pageNum := 0
		err = svc.ListObjectsPages(params, func(page *s3.ListObjectsOutput, lastPage bool) bool {
			for _, item := range page.Contents {
				size += *item.Size
			}
			pageNum++
			return true
		})
	} else if *flagProvider == PROVIDER_GS {
		protocol = PROTO_GS
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			writeCheckOutput(ReturnCode(CRITICAL), fmt.Sprintf("Error initializing: %s", err), "")
		}
		bkt := client.Bucket(*flagBucket)
		it := bkt.Objects(ctx, nil)

		for {
			objAttrs, err := it.Next()
			if err == iterator.Done {
				err = nil
				break
			}
			if err != nil {
				break
			}
			tmpSize := objAttrs.Size
			if strings.HasPrefix(objAttrs.Name, *flagPrefix) {
				if *flagDebug {
					fmt.Printf("Using %s %d\n", objAttrs.Name, tmpSize)
				}
				size += tmpSize
			} else {
				if *flagDebug {
					fmt.Printf("Skipping %s\n", objAttrs.Name)
				}
			}
		}
		if *flagDebug {
			fmt.Printf("Total size in byte: %d\n", size)
		}
	} else {
		execError()
	}

	if err != nil {
		writeCheckOutput(ReturnCode(UNKNOWN), fmt.Sprintf("Unable to list contents of bucket %s://%s", protocol, *flagBucket), "")
	}

	if minBytesWarn > int64(-1) && size < minBytesWarn {
		writeCheckOutput(ReturnCode(WARNING), fmt.Sprintf("Contents too small: %s://%s/%s", protocol, *flagBucket, *flagPrefix), "")
	}
	if maxBytesWarn > int64(-1) && size > maxBytesWarn {
		writeCheckOutput(ReturnCode(WARNING), fmt.Sprintf("Contents too big: %s://%s/%s", protocol, *flagBucket, *flagPrefix), "")
	}
	if minBytesCrit > int64(-1) && size < minBytesCrit {
		writeCheckOutput(ReturnCode(CRITICAL), fmt.Sprintf("Contents too small: %s://%s/%s", protocol, *flagBucket, *flagPrefix), "")
	}
	if maxBytesCrit > int64(-1) && size > maxBytesCrit {
		writeCheckOutput(ReturnCode(CRITICAL), fmt.Sprintf("Contents too big: %s://%s/%s", protocol, *flagBucket, *flagPrefix), "")
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
