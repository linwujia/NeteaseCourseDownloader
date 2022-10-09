package main

import "github.com/chromedp/cdproto/network"

type EventLoginCookies struct {
	cookies []*network.Cookie
}

type EventCrawlerError struct {
	err error
	msg string
}

type EventCourseCrawlerCompleted struct {
	courses []*Course
}

type EventChaptersCrawler struct {
	Id       int64
	chapters []*Chapter
}
