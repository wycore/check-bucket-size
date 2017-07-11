# check-bucket-size
[![GitHub version](https://badge.fury.io/gh/wywygmbh%2Fcheck-bucket-size.svg)](https://badge.fury.io/gh/wywygmbh%2Fcheck-bucket-size)
[![Build Status](https://travis-ci.org/wywygmbh/check-bucket-size.svg?branch=master)](https://travis-ci.org/wywygmbh/check-bucket-size)
[![Go Report](https://goreportcard.com/badge/github.com/wywygmbh/check-bucket-size)](https://goreportcard.com/report/github.com/wywygmbh/check-bucket-size)

Check contents of an S3 (or Google Cloud Storage) bucket or parts of it (minimum size and maximum size).

Compatible with Icinga, Nagios, Sensu, ... It uses the common exit codes.

## Motivation

This was originally created to check if a cassandra backup was successfully uploaded to an S3 bucket.

## Example

```
# check if s3://my-bucket/prod/my_cluster/20170402HHMMSS/ contains >= 100GB of data

$ ./check-bucket-size -provider s3 -bucket my-bucket -prefix prod/my_cluster/20170402 \
  -min-crit 100G -min-warn 120G

# check if gs://my-bucket/prod/my_cluster contains >= 100GB of data

$ ./check-bucket-size -provider gs -bucket my-bucket -prefix prod/my_cluster \
  -min-crit 100G -min-warn 120G
```

The `prefix` parameter is deliberately dumb, if you need to use some date arithmetic, you can
use existing methods like `$(date +"%Y%m%d" -d "last Sunday")`.
 
## Usage

    Usage of ./check-bucket-size:
      -bucket string
        bucket name (required)
      -max-crit 1234 / 1234k / 1234M / 1234G
        max-crit (default -1)
      -max-warn 1234 / 1234k / 1234M / 1234G
        max-bytes warn (default -1)
      -min-crit 1234 / 1234k / 1234M / 1234G
        min-crit (default -1)
      -min-warn 1234 / 1234k / 1234M / 1234G
        min-warn (default -1)
      -prefix string
        prefix in the bucket (optional)
      -provider string
        's3' for Amazon S3 or 'gs' for Google Cloud Storage


## Authentication

### Amazon S3

This check needs a `~/.aws/config` file in the following format:
```
[default]
region = eu-west-1
aws_access_key_id = ...
aws_secret_access_key = ...
```

### Google Cloud Storage

This check needs a "[Google Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials)" JSON file.

```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/something.json
./check-bucket-size ...
```

## How to build/test/etc

```bash
make test
make build
```

## License

Copyright 2017 wywy GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This code is being actively maintained by some fellow engineers at [wywy GmbH](http://wywy.com/).
