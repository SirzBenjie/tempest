package ashara

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Rest interface {
	Request(method string, route string, jsonPayload any) ([]byte, error)
	RequestWithFiles(method string, route string, jsonPayload any, files []*os.File) ([]byte, error)
}

type BaseRestClient struct {
	HTTPClient *http.Client
	Token      string
	MaxRetries uint8
	mu         sync.RWMutex
	lockedTo   time.Time
}

type rateLimitError struct {
	Message    string  `json:"message"`
	RetryAfter float32 `json:"retry_after"`
	Global     bool    `json:"global"`
}

func NewBaseRestClient(token string) Rest {
	t := token
	if !strings.HasPrefix(t, "Bot ") {
		t = "Bot " + t
	}

	return &BaseRestClient{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout: time.Second * 3,
			},
			Timeout: time.Second * 3,
		},
		Token:      t,
		MaxRetries: 3,
		lockedTo:   time.Time{},
	}
}

func (rest *BaseRestClient) Request(method string, route string, jsonPayload interface{}) ([]byte, error) {
	var body io.Reader
	if jsonPayload != nil {
		raw, err := json.Marshal(jsonPayload)
		if err != nil {
			return nil, errors.New("failed to parse provided payload (make sure it's in JSON format)")
		}

		body = bytes.NewReader(bytes.Replace(raw, requestSwapNullArray, requestSwapEmptyArray, -1))
	}

	if !rest.lockedTo.IsZero() {
		timeLeft := time.Until(rest.lockedTo)
		if timeLeft > 0 {
			time.Sleep(timeLeft)
		}
	}

	var i uint8 = 0
	for i < rest.MaxRetries {
		i++
		rest.mu.RLock()
		raw, err, finished := rest.handleRequest(method, route, body, CONTENT_TYPE_JSON)
		if finished {
			return raw, err
		}
		rest.mu.RUnlock()
		time.Sleep(time.Microsecond * time.Duration(250*i))
	}

	return nil, errors.New("failed to make http request in set limit of attempts to " + method + " :: " + route + " (check internet connection and/or app credentials)")
}

func (rest *BaseRestClient) RequestWithFiles(method string, route string, jsonPayload interface{}, files []*os.File) ([]byte, error) {
	if len(files) == 0 {
		return rest.Request(method, route, jsonPayload)
	}

	if !rest.lockedTo.IsZero() {
		timeLeft := time.Until(rest.lockedTo)
		if timeLeft > 0 {
			time.Sleep(timeLeft)
		}
	}

	var body *bytes.Buffer
	var writer *multipart.Writer
	if jsonPayload != nil {
		raw, err := json.Marshal(jsonPayload)
		if err != nil {
			return nil, errors.New("failed to parse provided payload (make sure it's in JSON format)")
		}

		body = bytes.NewBuffer(bytes.Replace(raw, requestSwapNullArray, requestSwapEmptyArray, -1))
		writer = multipart.NewWriter(body)
	}

	jsonPart, err := writer.CreatePart(partHeader(`form-data; name="payload_json"`, CONTENT_TYPE_JSON))
	if err != nil {
		return nil, errors.New("failed to create json body part in multipart payload: " + err.Error())
	}

	err = json.NewEncoder(jsonPart).Encode(jsonPayload)
	if err != nil {
		return nil, errors.New("failed to encode your json data into multipart payload: " + err.Error())
	}

	for itx, file := range files {
		num := strconv.Itoa(itx)

		stat, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to read statistics of file[%s]: %s", num, err)
		}

		filePart, err := writer.CreatePart(partHeader(fmt.Sprintf(`form-data; name="files[%s]"; filename="%s"`, num, stat.Name()), "application/octet-stream"))
		if err != nil {
			return nil, fmt.Errorf("failed to create body part in multipart for file[%s]: %s", num, err)
		}

		if _, err := io.Copy(filePart, file); err != nil {
			return nil, fmt.Errorf("failed to encode your \"%s\" file data into multipart payload: %s", file.Name(), err)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, errors.New("failed to close multipart payload: " + err.Error())
	}

	var i uint8 = 0
	for i < rest.MaxRetries {
		i++
		rest.mu.RLock()
		raw, err, finished := rest.handleRequest(method, route, body, writer.FormDataContentType())
		if finished {
			return raw, err
		}
		rest.mu.RUnlock()
		time.Sleep(time.Microsecond * time.Duration(250*i))
	}

	return nil, errors.New("failed to make http request 3 times to " + method + " :: " + route + " (check internet connection and/or app credentials)")
}

func (rest *BaseRestClient) handleRequest(method string, route string, payload io.Reader, contentType string) ([]byte, error, bool) {
	request, err := http.NewRequest(method, DISCORD_API_URL+route, payload)
	if err != nil {
		return nil, errors.New("failed to initialize new request: " + err.Error()), false
	}

	request.Header.Add("Content-Type", contentType)
	request.Header.Add("User-Agent", USER_AGENT)
	request.Header.Add("Authorization", rest.Token)

	res, err := rest.HTTPClient.Do(request)
	if err != nil {
		return nil, errors.New("failed to process request: " + err.Error()), false
	}

	if res.StatusCode == 204 {
		return nil, nil, true
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("failed to parse response body (json): " + err.Error()), true
	}

	if res.StatusCode == 429 {
		rateErr := rateLimitError{}
		json.Unmarshal(body, &rateErr)

		rest.mu.Lock()
		timeLeft := time.Now().Add(time.Second * time.Duration(rateErr.RetryAfter+5))
		rest.lockedTo = timeLeft
		rest.mu.Unlock()

		time.Sleep(time.Until(timeLeft))

		rest.mu.Lock()
		rest.lockedTo = time.Time{}
		rest.mu.Unlock()
		return nil, errors.New("rate limit"), false
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(res.Status + " :: " + string(body)), true
	}

	return body, nil, true
}

func partHeader(contentDisposition string, contentType string) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Disposition": []string{contentDisposition},
		"Content-Type":        []string{contentType},
	}
}
