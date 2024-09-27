package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

func main() {
	var (
		logGroupName string
	)
	flag.StringVar(&logGroupName, "log-group", "", "log group to ingest")
	flag.Parse()

	if logGroupName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := cloudwatchlogs.NewFromConfig(cfg)

	logStreams, err := fetchLogStreams(svc, logGroupName)
	if err != nil {
		log.Fatalf("failed to fetch log streams: %v", err)
	}

	for _, stream := range logStreams {
		err := fetchLogEvents(svc, logGroupName, *stream.LogStreamName)
		if err != nil {
			log.Printf("failed to fetch log events for stream %s: %v", *stream.LogStreamName, err)
			return
		}
	}
}

func fetchLogStreams(svc *cloudwatchlogs.Client, logGroupName string) ([]types.LogStream, error) {
	var allStreams []types.LogStream
	nextToken := ""

	for {
		input := &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(logGroupName),
		}

		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		result, err := svc.DescribeLogStreams(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		allStreams = append(allStreams, result.LogStreams...)

		if result.NextToken == nil {
			break
		}

		nextToken = *result.NextToken
	}

	return allStreams, nil
}

func fetchLogEvents(svc *cloudwatchlogs.Client, logGroupName, logStreamName string) error {
	nextToken := ""
	messages := []map[string]any{}
	logStreamNameSplit := strings.Split(logStreamName, "/")
	logStreamWithoutRandom := strings.Join(logStreamNameSplit[:len(logStreamNameSplit)-1], "/")

	for {
		input := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(logGroupName),
			LogStreamName: aws.String(logStreamName),
			StartFromHead: aws.Bool(true),
		}

		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		result, err := svc.GetLogEvents(context.TODO(), input)
		if err != nil {
			return err
		}

		for _, event := range result.Events {
			seconds := float64(*event.Timestamp / 1000)
			microseconds := float64(*event.Timestamp%1000) * 1000
			messages = append(messages, map[string]any{
				"date":       seconds + (microseconds / 1e6),
				"log":        *event.Message,
				"log-group":  logGroupName,
				"log-stream": logStreamWithoutRandom,
			})
		}

		if result.NextForwardToken == nil || nextToken == *result.NextForwardToken {
			break
		}

		nextToken = *result.NextForwardToken
	}

	if len(messages) == 0 {
		return nil
	}

	out, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost/api/observability/ingestion/json", "image/jpeg", bytes.NewBuffer(out))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code is not 200")
	}

	fmt.Printf("Ingested log-group %s, stream %s: %d messages\n", logGroupName, logStreamWithoutRandom, len(messages))

	return nil
}
