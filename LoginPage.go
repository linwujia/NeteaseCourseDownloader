package main

import (
	"NeteaseCourseDownloader/chrome"
	"github.com/golang/glog"
	"github.com/tebeka/selenium"
	"net/http"
	"strings"
	"time"
)

const (
	LoginUrl = "https://study.163.com/member/login.htm"
)

type LoginPage struct {
	service *selenium.Service
}

func NewLoginPage() *LoginPage {
	page := new(LoginPage)
	return page
}

func (p *LoginPage) Show(loginDriver chrome.IWebDriver) (cookies []*http.Cookie, err error) {

	err = loginDriver.Get(LoginUrl)
	if err != nil {
		glog.Error("get login page failed", err.Error())
		return
	}

	loginDriver.Wait(func(wd selenium.WebDriver) (bool, error) {
		title, err2 := loginDriver.Title()
		if err2 != nil {
			return false, err2
		}
		return strings.EqualFold(title, "网易云课堂 - 悄悄变强大"), nil
	})

	cookiesTemp, err := loginDriver.GetCookies()
	if err != nil {
		return
	}

	cookies = make([]*http.Cookie, len(cookiesTemp))
	for i, cookie := range cookiesTemp {
		cookies[i] = &http.Cookie{
			Name:       cookie.Name,
			Value:      cookie.Value,
			Path:       cookie.Path,
			Domain:     cookie.Domain,
			Expires:    time.Unix(int64(cookie.Expiry), 0),
			RawExpires: "",
			MaxAge:     0,
			Secure:     cookie.Secure,
			HttpOnly:   false,
			SameSite:   0,
			Raw:        "",
			Unparsed:   nil,
		}
	}

	return
}
