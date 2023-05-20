package contrib

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/http"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/proto"
)

type HTTPRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

type HTTPResponse struct {
	URL        string            `json:"url"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"statusCode"`
}

func NewHTTPActivity() activity.WorkerActivity {
	return NewHTTPActivityWithConfig(64, 10, 30)
}

func NewHTTPActivityWithConfig(maxConcurrentRequests, maxConnsPerHost, maxTimeoutSec int) activity.WorkerActivity {
	description := "An activity to fire http requests on the worker node"
	httpClient := http.NewBuilder().Id("http-activity").MaxConcurrentRequests(maxConcurrentRequests).MaxConnsPerHost(maxConnsPerHost).MaxQueueSize(runtime.NumCPU() * 512).TimeoutSec(maxTimeoutSec).Build()
	return activity.NewWorkerActivity(&proto.Activity{
		Name:        "http activity",
		Description: &description,
	}, makeActivityHandler(httpClient))
}

func makeActivityHandler(httpClient http.HTTPClient) activity.ActivityHandler {
	return func(ctx context.Context, logger logging.Logger, input []byte) (output []byte, err error) {
		var inputRequest HTTPRequest
		if err := json.Unmarshal(input, &inputRequest); err != nil {
			return nil, err
		}
		logger.Infof(ctx, "process http request %v", inputRequest)
		if inputRequest.Method == "" {
			return nil, errors.Error("method is required")
		}
		if inputRequest.URL == "" {
			return nil, errors.Error("url is required")
		}
		if !strings.HasPrefix(inputRequest.URL, "http://") || !strings.HasPrefix(inputRequest.URL, "https://") {
			inputRequest.URL = "http://" + inputRequest.URL
		}
		requestBuilder := http.NewRequestBuilder().URL(inputRequest.URL).Method(strings.ToUpper(inputRequest.Method))
		if inputRequest.Headers != nil {
			headerBuilder := http.NewHeaderMaker()
			for k, v := range inputRequest.Headers {
				headerBuilder.Set(k, v)
			}
			requestBuilder.Header(headerBuilder.Make())
		}
		if strings.ToUpper(inputRequest.Method) != "GET" && inputRequest.Body != "" {
			requestBuilder.StringBody(inputRequest.Body)
		}
		return fireHTTPRequest(httpClient, requestBuilder.Build())
	}
}

func fireHTTPRequest(httpClient http.HTTPClient, request *http.Request) ([]byte, error) {
	resp := httpClient.Request(request)
	return serializeResponse(resp)
}

func serializeResponse(response *http.Response) ([]byte, error) {
	uri := response.URI
	body := response.Body
	headers := response.Header
	statusCode := response.Code
	outputHeaders := make(map[string]string)
	for k, v := range headers {
		outputHeaders[k] = v[0]
	}
	outputResponse := HTTPResponse{
		URL:        uri,
		Body:       body,
		Headers:    outputHeaders,
		StatusCode: statusCode,
	}
	return json.Marshal(outputResponse)
}
