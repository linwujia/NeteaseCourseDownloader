package main

import (
	"NeteaseCourseDownloader/chrome"
	"github.com/chromedp/cdproto/network"
	"github.com/golang/glog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type CourseManager struct {
	url         string
	accountUser *AccountUser
	crawler     ICourseCrawler
	eventChanel chan interface{}

	cookies []*http.Cookie
	courses []*Course
}

var wg sync.WaitGroup

var WaitCookies = make(chan interface{})
var WaitCrawler = make(chan interface{})

func NewCourseManager(url string) *CourseManager {
	manager := new(CourseManager)
	manager.url = url
	manager.accountUser = NewAccountUser(manager)
	manager.crawler = NewHttpCourseCrawler(manager, url)
	manager.eventChanel = make(chan interface{})
	return manager
}

func (manager *CourseManager) Init() {
	go manager.handleEvent()
	manager.crawler.Init()
}

func (manager *CourseManager) handleEvent() {
	for {
		event := <-manager.eventChanel
		switch e := event.(type) {
		case EventLoginCookies:
			manager.ParseLoginCookies(e)
		case EventCourseCrawlerCompleted:
			manager.handleCourseCrawlerCompleted(e)
		case EventChaptersCrawler:
			manager.handleChapterCrawler(e)
		default:
			glog.Error("unknown event: ", event)
		}
	}
}

func (manager CourseManager) SendEvent(event interface{}) {
	manager.eventChanel <- event
}

func (manager *CourseManager) Run() {

	if err := manager.accountUser.LoginIfNeed(); err != nil {
		glog.Fatal("登录失败", err)
	}

	<-WaitCookies
	manager.crawler.CourseCrawler()

	<-WaitCrawler

	/*if !isLogin {
		glog.Info("is not login, please login first")
		cookies, err := manager.showLoginPage()
		if err != nil {
			glog.Fatal("login error ", err)
		}

		chapters, err := manager.getChapterList(cookies)
		if err != nil {
			glog.Fatal("login error ", err)
		}

		sessionUrl := "https://course.study.163.com/j/cp/lecture/front/getList.json"
		for _, chapter := range chapters {
			wg.Add(1)
			go manager.getChapterSessions(cookies, sessionUrl, chapter)
		}

		wg.Wait()
	}*/
}

func (manager *CourseManager) Stop() {
	chrome.Client.Close()
}

func (manager CourseManager) ParseLoginCookies(event EventLoginCookies) {
	cookies := make([]*http.Cookie, len(event.cookies))
	for i, cookie := range event.cookies {
		cookies[i] = &http.Cookie{
			Name:       cookie.Name,
			Value:      cookie.Value,
			Path:       cookie.Path,
			Domain:     cookie.Domain,
			Expires:    time.Unix(int64(cookie.Expires), 0),
			RawExpires: "",
			MaxAge:     0,
			Secure:     cookie.Secure,
			HttpOnly:   cookie.HTTPOnly,
			SameSite:   manager.getSameSite(cookie),
			Raw:        "",
			Unparsed:   nil,
		}
	}

	manager.cookies = cookies
	manager.crawler.SetCookies(cookies)
	WaitCookies <- struct{}{}
}

func (manager *CourseManager) handleCourseCrawlerCompleted(event EventCourseCrawlerCompleted) {
	manager.courses = event.courses
	for _, course := range event.courses {
		go manager.crawler.ChapterCrawler(course)
	}
}

func (manager CourseManager) handleChapterCrawler(event EventChaptersCrawler) {
	for _, course := range manager.courses {
		if course.Id == event.Id {
			course.chapters = event.chapters
			break
		}
	}

	for _, chapter := range event.chapters {
		wg.Add(1)
		go manager.crawler.ContentCrawler(strconv.FormatInt(event.Id, 10), chapter)
	}

	defer func() {
		WaitCrawler <- struct{}{}
	}()

	wg.Wait()
}

func (manager *CourseManager) getSameSite(cookie *network.Cookie) http.SameSite {
	sameSiteMap := map[network.CookieSameSite]http.SameSite{
		network.CookieSameSiteNone:   http.SameSiteNoneMode,
		network.CookieSameSiteLax:    http.SameSiteLaxMode,
		network.CookieSameSiteStrict: http.SameSiteStrictMode,
	}

	if site, ok := sameSiteMap[cookie.SameSite]; ok {
		return site
	}

	return http.SameSiteDefaultMode
}
