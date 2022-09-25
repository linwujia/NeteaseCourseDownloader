package main

import (
	"github.com/golang/glog"
	"sync"
)

type CourseManager struct {
	url          string
	chromeClient *ChromeDpClient
	accountUser  *AccountUser
	crawler      *CourseCrawler
}

var wg sync.WaitGroup

func NewCourseManager(url string) *CourseManager {
	manager := new(CourseManager)
	manager.url = url
	manager.chromeClient = NewChromeDpClient()
	manager.accountUser = NewAccountUser(manager.chromeClient)
	manager.crawler = NewCourseCrawler(manager.chromeClient)
	return manager
}

func (manager *CourseManager) Init() {
	manager.chromeClient.Init()
	manager.crawler.Init()
}

func (manager *CourseManager) Run() {

	if err := manager.accountUser.LoginIfNeed(); err != nil {
		glog.Fatal("登录失败", err)
	}

	manager.crawler.CrawlerCourse()
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
	manager.chromeClient.Close()
}
