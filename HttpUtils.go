package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	url1 "net/url"
)

var DefaultHeaders = map[string]string{
	//"accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	//"accept-encoding": "gzip, deflate, br",
	"referer":    "https://study.163.com/",
	"sec-ch-ua":  "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"100\", \"Google Chrome\";v=\"100\"",
	"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36",
}

func HttpRequest(method string, url string, header map[string]string, init func(r *http.Request), requestBody map[string]string) (result []byte, err error) {

	values := url1.Values{}
	if requestBody != nil {
		for key, value := range requestBody {
			values[key] = []string{value}
		}
	}

	request, err := http.NewRequest(method, url, bytes.NewReader([]byte(values.Encode())))
	if err != nil {
		fmt.Errorf("error %e", err)
		return
	}

	AddHeaders(request, header)

	if init != nil {
		init(request)
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Errorf("error %e", err)
		return
	}

	defer response.Body.Close()

	result, err = io.ReadAll(response.Body)
	return
}

func HttpGet(url string, header map[string]string, init func(r *http.Request)) (result []byte, err error) {
	method := "GET"
	return HttpRequest(method, url, header, init, nil)
}

func HttpPost(url string, header map[string]string, init func(r *http.Request), formData map[string]string) (result []byte, err error) {
	method := "POST"
	return HttpRequest(method, url, header, init, formData)
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
