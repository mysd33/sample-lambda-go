package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	var response events.SQSEventResponse

	for _, v := range event.Records {
		body := v.Body
		log.Printf("Message: %s", body)
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
