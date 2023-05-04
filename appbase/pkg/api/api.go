package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

func ParsePostRequest(req events.APIGatewayProxyRequest, v interface{}) error {
	if req.HTTPMethod != http.MethodPost {
		return errors.Errorf("use POST request")
	}
	err := json.Unmarshal([]byte(req.Body), v)
	if err != nil {
		return errors.Wrapf(err, "failed to parse request")
	}
	return nil
}

func OkResponse(result string) (events.APIGatewayProxyResponse, error) {
	return response(
		http.StatusOK,
		result), nil
}

func ErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return response(
		http.StatusBadRequest,
		errorResponseBody(err.Error())), nil
}

func response(code int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       body,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func errorResponseBody(msg string) string {
	return fmt.Sprintf("{\"message\":\"%s\"}", msg)
}
