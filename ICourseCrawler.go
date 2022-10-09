package main

import "net/http"

type ICourseCrawler interface {
	Init()
	SetCookies(cookies []*http.Cookie)
	CourseCrawler()                                   //爬取课程
	ChapterCrawler(course *Course)                    //爬取章
	ContentCrawler(courseId string, chapter *Chapter) //爬取节，或者内容即没有节时的列表
}
