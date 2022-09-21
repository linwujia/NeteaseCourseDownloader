package myhttp

import (
	"errors"
	"io"
	"net/http"
)

type ResponseMsg struct {
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Ok      bool                   `json:"ok"`
}

func HttpGet(url string, before func(r *http.Request)) (response *http.Response, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	if before != nil {
		before(request)
	}

	response, err = http.DefaultClient.Do(request)
	return
}

func HttpGetInterface(url string, before func(r *http.Request), convert func(*http.Response) (*ResponseMsg, error)) (result *ResponseMsg, err error) {
	response, err := HttpGet(url, before)
	if err != nil {
		return
	}

	result, err = convert(response)
	return
}

func HttpGetBytes(url string, before func(r *http.Request)) (bytes []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	if before != nil {
		before(request)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}

	bytes, err = ReadBytes(response)
	return
}

func ReadBytes(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.New("response is nil")
	}

	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	return bytes, err
}

func AddCookies(r *http.Request, cookies []*http.Cookie) {
	if cookies == nil || len(cookies) == 0 {
		return
	}
	if r == nil {
		return
	}
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
}

func AddHeaders(r *http.Request, headers map[string]string) {
	if headers == nil || len(headers) == 0 {
		return
	}
	if r == nil {
		return
	}

	for key, value := range headers {
		r.Header.Add(key, value)
	}
}

func AddHeader(r *http.Request, key, value string) {
	if r == nil {
		return
	}

	r.Header.Add(key, value)
}
