package lubelogger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Construct a full request URL by combining the supplied endpoint with the
// api_url configuration value

func endpointURL(endpoint string) string {
	return fmt.Sprintf("%s/%s", api_uri, endpoint)
}

// Value for the authorization header expected by th LubeLogger API

func authorizationHeader() string {
	return authorization
}

// LubeLogger endpoint drop-in replacement for http.Get()

func GetEndpoint(endpoint string) (*http.Response, error) {
	requestURL := endpointURL(endpoint)

	log.WithFields(log.Fields{
		"endpoint":   endpoint,
		"requestURL": requestURL,
	}).Trace("GetEndpoint called")

	apiRequest, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("GetEndpoint NewRequest: %w", err)
	}

	apiRequest.Header.Add("Authorization", authorizationHeader())

	apiResponse, err := http.DefaultClient.Do(apiRequest)

	log.WithFields(log.Fields{
		"endpoint":      endpoint,
		"statusCode":    apiResponse.StatusCode,
		"proto":         apiResponse.Proto,
		"contentLength": apiResponse.ContentLength,
	}).Debug("GetEndpoint apiResponse")

	return apiResponse, err
}

// Wrapped Get for standardized call of API GET endpoints

func APIGet(endpoint string) ([]byte, error) {
	apiResponse, err := GetEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("APIGet GetEndpoint: %w", err)
	}

	responseBody, err := io.ReadAll(apiResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("reading responseBody: %w", err)
	}

	log.WithFields(log.Fields{
		"responseBytes": len(responseBody),
	}).Debug("LubeLogger APIGet")

	return responseBody, nil
}

// LubeLogger endpoint drop-in replacement for http.PostForm()

func PostFormEndpoint(endpoint string, data url.Values) (*http.Response, error) {
	requestURL := endpointURL(endpoint)

	log.WithFields(log.Fields{
		"endpoint":   endpoint,
		"requestURL": requestURL,
		"data":       data,
	}).Trace("PostFormEndpoint called")

	requestBody := strings.NewReader(data.Encode())

	apiRequest, err := http.NewRequest(http.MethodPost, requestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("PostFormEndpoint NewRequest: %w", err)
	}

	apiRequest.Header.Add("Authorization", authorizationHeader())
	apiRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	apiResponse, err := http.DefaultClient.Do(apiRequest)

	log.WithFields(log.Fields{
		"endpoint":      endpoint,
		"statusCode":    apiResponse.StatusCode,
		"proto":         apiResponse.Proto,
		"contentLength": apiResponse.ContentLength,
	}).Debug("PostFormEndpoint apiResponse")

	return apiResponse, err
}

// Wrapped PostForm for standardized call of API POST endpoints

func APIPostForm(endpoint string, data url.Values) (response PostResponse, err error) {
	apiResponse, err := PostFormEndpoint(endpoint, data)
	if err != nil {
		return PostResponse{}, fmt.Errorf("APIPostForm PostFormEndpoint: %w", err)
	}

	responseBody, err := io.ReadAll(apiResponse.Body)
	if err != nil {
		return PostResponse{}, fmt.Errorf("APIPostForm reading responseBody: %w", err)
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return PostResponse{}, fmt.Errorf("unmarshalling json: %w", err)
	}

	log.WithFields(log.Fields{
		"success":    response.Success,
		"message":    response.Message,
		"status":     apiResponse.StatusCode,
		"formBytes":  len(data.Encode()),
		"formFields": len(data),
	}).Debug("LubeLogger APIPostForm")

	if !response.Success {
		return response, fmt.Errorf("post: %s", response.Message)
	}

	return response, nil
}
