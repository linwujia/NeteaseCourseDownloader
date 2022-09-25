package main

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/golang/glog"
	"strings"
)

type CourseCrawler struct {
	chromeDpClient *ChromeDpClient
	courseIndexUrl string
	courseChapters []*CourseChapter

	eventChanel chan interface{}

	waitChanel chan interface{}
}

func NewCourseCrawler(chromeDpClient *ChromeDpClient) *CourseCrawler {
	crawler := new(CourseCrawler)
	crawler.chromeDpClient = chromeDpClient
	crawler.courseIndexUrl = "https://course.study.163.com/480000005355162/learning"
	crawler.eventChanel = make(chan interface{})
	crawler.waitChanel = make(chan interface{})
	return crawler
}

func (crawler CourseCrawler) Init() {
	go crawler.handleEvent()
}

func (crawler *CourseCrawler) CrawlerCourse() error {
	return crawler.chromeDpClient.Run(crawler.crawlerCourse())
}

func (crawler *CourseCrawler) crawlerCourse() chromedp.Tasks {
	return chromedp.Tasks{
		// 1 爬取课程 章目
		crawler.crawlerChapter(),
		// 2 爬取课程 节目
		crawler.crawlerSessions(),
		/*// 3 保持cookies
		user.saveCookies(),*/
	}
}

func (crawler *CourseCrawler) crawlerChapter() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if err = chromedp.Navigate(crawler.courseIndexUrl).Do(ctx); err != nil {
			return
		}

		if err = chromedp.WaitVisible("#j-coursearrangeuiwrap > div > div > div").Do(ctx); err != nil {
			return
		}

		var chapter string
		if err = chromedp.OuterHTML("#j-coursearrangeuiwrap > div > div", &chapter).Do(ctx); err != nil {
			return
		}

		glog.Info("chapter ", chapter)

		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(chapter)); err != nil {
			return err
		} else {
			doc.Find("div.term").Each(func(i int, selection *goquery.Selection) {

				var chapter CourseChapter
				chapter.Title = selection.Find(`span[class="f-fl name f-thide"]`).Text()
				glog.Info(chapter.Title)

				chapter.Link, _ = selection.Find("a").First().Attr("href")
				chapter.Link = "https:" + chapter.Link
				crawler.courseChapters = append(crawler.courseChapters, &chapter)
			})
		}

		return
	}
}

func (crawler *CourseCrawler) crawlerSessions() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for _, chapter := range crawler.courseChapters {
			err = crawler.crawlerSingleSession(ctx, chapter)
			if err != nil {
				return err
			}
		}
		return
	}
}

func (crawler *CourseCrawler) crawlerSingleSession(ctx context.Context, chapter *CourseChapter) (err error) {

	err = chromedp.Navigate(chapter.Link).Do(ctx)
	if err != nil {
		glog.Error(err)
		return
	}

	if err = chromedp.WaitVisible("#j-chapteruiwrap > div > div > div > ul:nth-child(1)").Do(ctx); err != nil {
		glog.Error(err)
		return
	}

	var sessions string
	if err = chromedp.OuterHTML("#j-chapteruiwrap > div > div > div", &sessions).Do(ctx); err != nil {
		glog.Error(err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(sessions))
	if err != nil {
		glog.Error(err)
		return
	}

	doc.Find(`/ul[@class="ux-learn-lecture-tree_list"]`).Each(func(i int, selection *goquery.Selection) {
		var session CourseSession
		session.Title = selection.Find("li > div div span").Text()
		selection.Find("div.ux-learn-lecture-tree").Each(func(i int, selection *goquery.Selection) {
			var child Child
			child.Title = selection.Find("li > div.ux-learn-lecture-tree > ul:nth-child(1) > li > div.ux-learn-lecture-tree_item > div > span").Text()

			child.Children = append(child.Children, &child)
		})

		chapter.Sessions = append(chapter.Sessions, &session)
	})

	glog.Info(sessions)
	return
}

func (crawler *CourseCrawler) handleEvent() {
	for {
		event := <-crawler.eventChanel
		switch e := event.(type) {
		case EventAddChapter:
			crawler.courseChapters = append(crawler.courseChapters, e.Chapter)
		default:
			glog.Error("unknown event: ", event)
		}
	}
}

func (crawler CourseCrawler) SendEvent(event interface{}) {
	crawler.eventChanel <- event
}

func (crawler *CourseCrawler) Close() {
	close(crawler.eventChanel)
}
