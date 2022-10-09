package main

import (
	"NeteaseCourseDownloader/chrome"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/golang/glog"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

var chapterUrl = "https://course.study.163.com/j/cp/lecture/front/getList.json"

type HttpCourseCrawler struct {
	url     string
	cookies []*http.Cookie
	manager *CourseManager
}

func NewHttpCourseCrawler(manager *CourseManager, url string) *HttpCourseCrawler {
	crawler := new(HttpCourseCrawler)
	crawler.url = url
	crawler.manager = manager
	return crawler
}

func (crawler *HttpCourseCrawler) Init() {
	glog.Info("HttpCourseCrawler Init")
}

func (crawler *HttpCourseCrawler) SetCookies(cookies []*http.Cookie) {
	crawler.cookies = cookies
}

func (crawler *HttpCourseCrawler) CourseCrawler() {
	/*header := make(map[string]string)
	for key, value := range DefaultHeaders {
		header[key] = value
	}

	header["Content-Type"] = "application/x-www-form-urlencoded"

	formData := make(map[string]string)
	formData["termId"] = "480000005355162"
	formData["preview"] = "0"

	courseJson, err := HttpPost("https://course.study.163.com/j/cp/getCompositeRelList.json", header, crawler.AddCookies, formData)

	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "course crawler request error")
		return
	}

	glog.Info(string(courseJson))

	var response Response
	err = json.Unmarshal(courseJson, &response)
	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "course crawler response unmarshal error")
		return
	}

	result, err := json.Marshal(response.Result)
	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "course crawler response unmarshal error")
		return
	}

	var courses []*Course

	err = json.Unmarshal(result, &courses)
	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "course crawler response course unmarshal error")
		return
	}

	crawler.manager.SendEvent(EventCourseCrawlerCompleted{courses: courses})*/

	crawler.downloadM3U8File("https://course.study.163.com/480000005356155/lecture-480000037092978")
}

func (crawler *HttpCourseCrawler) ChapterCrawler(course *Course) {
	header := make(map[string]string)
	for key, value := range DefaultHeaders {
		header[key] = value
	}

	header["Content-Type"] = "application/x-www-form-urlencoded"

	formData := make(map[string]string)
	formData["termId"] = strconv.FormatInt(course.Id, 10)
	formData["preview"] = "0"

	chapterJson, err := HttpPost(chapterUrl, header, crawler.AddCookies, formData)

	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "chapter crawler request err")
		return
	}

	var response Response
	err = json.Unmarshal(chapterJson, &response)
	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "chapter crawler http response Unmarshal err")
		return
	}

	result, err := json.Marshal(response.Result)
	if err != nil {
		glog.Error(err)
		crawler.SendError(err, "chapter crawler http response Marshal err")
		return
	}

	var list CataLogList
	if err = json.Unmarshal(result, &list); err != nil {
		glog.Error(err)
		crawler.SendError(err, "chapter crawler http response Unmarshal CataLogList err")
		return
	}

	crawler.manager.SendEvent(EventChaptersCrawler{
		Id:       course.Id,
		chapters: list.CataLogList,
	})
}

func (crawler *HttpCourseCrawler) ContentCrawler(courseId string, chapter *Chapter) {
	defer wg.Done()
	for _, child := range chapter.Children {
		wg.Add(1)
		go crawler.CrawlerVideo(courseId, child)
	}
}

func (crawler *HttpCourseCrawler) CrawlerVideo(courseId string, child *Child) {
	defer wg.Done()
	if len(child.Children) > 0 {
		for _, c := range child.Children {
			wg.Add(1)
			go crawler.CrawlerVideo(courseId, c)
		}
		return
	}

	url := fmt.Sprintf("https://course.study.163.com/%s/lecture-%d", courseId, child.Id)
	crawler.downloadM3U8File(url)
}

func (crawler *HttpCourseCrawler) AddCookies(r *http.Request) {
	AddCookies(r, crawler.cookies)
}

func (crawler *HttpCourseCrawler) SendError(err error, msg string) {
	crawler.manager.SendEvent(EventCrawlerError{
		err: err,
		msg: msg,
	})
}

func (crawler *HttpCourseCrawler) downloadM3U8File(url string) {
	m3u8Done := make(chan bool)
	keyDone := make(chan bool)

	var m3u8RequestID network.RequestID
	var keyRequestID network.RequestID

	m3u8Reg := regexp.MustCompile(`^https:.+\.m3u8.+`)
	keyReg := regexp.MustCompile(`^https:.+/key\?id=.+&token=.+`)

	var m3u8Url string

	chromedp.ListenTarget(chrome.Client.Context, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			log.Printf("EventRequestWillBeSent: %v: %v", ev.RequestID, ev.Request.URL)
			if m3u8Reg.MatchString(ev.Request.URL) {
				m3u8Url = ev.Request.URL
				m3u8RequestID = ev.RequestID
			} else if keyReg.MatchString(ev.Request.URL) {
				keyRequestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			log.Printf("EventLoadingFinished: %v", ev.RequestID)
			if ev.RequestID == m3u8RequestID {
				close(m3u8Done)
			} else if ev.RequestID == keyRequestID {
				close(keyDone)
			}
		}
	})

	if err := chrome.Client.Run(chromedp.Navigate(url)); err != nil {
		glog.Error(err)
		crawler.SendError(err, fmt.Sprintf("Navigate to %s error", url))
	}

	// This will block until the chromedp listener closes the channel
	<-m3u8Done
	<-keyDone
	// get the downloaded bytes for the request id
	var m3u8Buf []byte
	var keyBuf []byte
	if err := chrome.Client.Run(chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		m3u8Buf, err = network.GetResponseBody(m3u8RequestID).Do(ctx)
		if err != nil {
			return err
		}

		keyBuf, err = network.GetResponseBody(keyRequestID).Do(ctx)
		return err
	})); err != nil {
		log.Fatal(err)
	}

	NewM3U8Downloader("", m3u8Url, keyBuf, string(m3u8Buf), "", crawler.cookies).Download()
	// write the file to disk - since we hold the bytes we dictate the name and
	// location
	if err := ioutil.WriteFile("download.m3u8", m3u8Buf, 0644); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("download.key", keyBuf, 0644); err != nil {
		log.Fatal(err)
	}
	log.Print("wrote download.png")
}

func (crawler HttpCourseCrawler) downloadM3U8Tasks(url string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
	}
}
