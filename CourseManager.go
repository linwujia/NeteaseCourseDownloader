package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SeleniumPath = `.\executable\chromedriver`
	Port         = 9515
)

type CourseManager struct {
	url       string
	loginPage *LoginPage
}

var wg sync.WaitGroup

func NewCourseManager(url string) *CourseManager {
	manager := new(CourseManager)
	manager.url = url
	manager.loginPage = NewLoginPage()
	return manager
}

func (manager *CourseManager) Run() {
	isLogin := manager.isLogin()
	if !isLogin {
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
	}
}

func (manager *CourseManager) Stop() {

}

func (manager *CourseManager) isLogin() bool {
	html, err := HttpGet(manager.url, DefaultHeaders, nil)
	if err != nil {
		glog.Fatal("check login state err ", err)
	}

	return !strings.Contains(string(html), `id="j-nav-login"`)
}

func (manager *CourseManager) showLoginPage() (cookies []*http.Cookie, err error) {
	webDriver, err := Init(SeleniumPath, Port)
	if err != nil {
		err = fmt.Errorf("启动浏览器失败，%v", err)
		glog.Error(err)
		return
	}

	defer webDriver.Quit()
	return manager.loginPage.Show(webDriver)
}

func (manager *CourseManager) getChapterList(cookies []*http.Cookie) (chapters []CourseChapter, err error) {
	header := make(map[string]string)
	for key, value := range DefaultHeaders {
		header[key] = value
	}

	header["Content-Type"] = "application/x-www-form-urlencoded"

	formData := make(map[string]string)
	formData["termId"] = "480000005355162"
	formData["preview"] = "0"

	courseJson, err := HttpPost("https://course.study.163.com/j/cp/getCompositeRelList.json", header, func(r *http.Request) {
		AddCookies(r, cookies)
	}, formData)

	if err != nil {
		glog.Error(err)
		return
	}

	glog.Info(string(courseJson))

	var response Response
	err = json.Unmarshal(courseJson, &response)
	if err != nil {
		glog.Error(err)
		return
	}

	result, err := json.Marshal(response.Result)
	if err != nil {
		glog.Error(err)
		return
	}
	err = json.Unmarshal(result, &chapters)
	return
}

func (manager CourseManager) getChapterSessions(cookies []*http.Cookie, sessionUrl string, chapter CourseChapter) {
	defer wg.Done()

	header := make(map[string]string)
	for key, value := range DefaultHeaders {
		header[key] = value
	}

	header["Content-Type"] = "application/x-www-form-urlencoded"

	formData := make(map[string]string)
	formData["termId"] = strconv.FormatInt(chapter.Id, 10)
	formData["preview"] = "0"

	sessionJson, err := HttpPost(sessionUrl, header, func(r *http.Request) {
		AddCookies(r, cookies)
	}, formData)

	if err != nil {
		glog.Error(err)
		return
	}

	var response Response
	err = json.Unmarshal(sessionJson, &response)
	if err != nil {
		glog.Error(err)
		return
	}

	glog.Info(string(sessionJson))

	result, err := json.Marshal(response.Result)
	if err != nil {
		glog.Error(err)
		return
	}

	var list CataLogList
	if err = json.Unmarshal(result, &list); err != nil {
		glog.Error(err)
		return
	}

	for _, session := range list.CataLogList {
		manager.PrepareDownloader(chapter, session, cookies)
	}

	//请求m3u8文件  GET
	//"https://jdvodluwytr3t.vod.126.net/jdvodluwytr3t/nos/ept/hls/2019/08/19/1214987821_cbcee5fc7779461b89e4438772432a51_eshd.m3u8?ak=7909bff134372bffca53cdc2c17adc27a4c38c6336120510aea1ae1790819de816f8b22fb2a630f5a9abc99ea0a26c1e56bac650a4922d83351e84b232884bf29da33a2fceb3a0d16b18820211b50feb04eedd32d103bfabab5d3a1999d53979c7ac434e1ab4f73d44ab4c497dbfe9fee6fc2ae2e15a00d5fd833ea365cf99f4f6c8f6721206d89b7e93d9c01fcec819&token=https%3A%2F%2Fvod.study.163.com%2Feds%2Fapi%2Fv1%2Fvod%2Fhls%2Fkey%3Fid%3D1214987821%26token%3D6ac4053d2860b0203f5b620870516b0acd41bac795822906c1e3c04b829a742140f715ba01304ca42e66b1779f9969e9f75214b0e65b3da892000a0db90c953bf407db0a04221a0a5b475e5c9df80311532c3f3c0e2e635e4a10a532daee0b18671e1bf5f8aa9769ad15bef447b747e37a42404d535868f94a5b4533c2e69801&t=1663767438599"
}

func (manager CourseManager) PrepareDownloader(chapter CourseChapter, session CourseSession, cookies []*http.Cookie) {

	for _, child := range session.Children {
		wg.Add(1)
		go manager.DownloadCourseVideo(chapter, child, cookies)
	}
}

func (manager *CourseManager) DownloadCourseVideo(chapter CourseChapter, child Child, cookies []*http.Cookie) {
	defer wg.Done()

	if len(child.Children) > 0 {
		for _, c := range child.Children {
			wg.Add(1)
			manager.DownloadCourseVideo(chapter, c, cookies)
		}
		return
	}

	//获取signature POST
	//"https://course.study.163.com/p/cp/lecture/front/getLectureResource.json"
	/*contentId:	1214457506
	contentType:1
	id:480000010460327
	nodeType:3
	preview:0
	productType:"study"
	scene-type-id:"front-3-480000010460327"
	termId:480000005354155*/

	header := make(map[string]string)
	for key, value := range DefaultHeaders {
		header[key] = value
	}

	header["Content-Type"] = "application/x-www-form-urlencoded"

	formData := make(map[string]string)
	formData["contentId"] = strconv.Itoa(child.Data.ContentId)
	formData["contentType"] = strconv.Itoa(child.Data.ContentType)
	formData["id"] = strconv.FormatInt(child.Id, 10)
	formData["nodeType"] = strconv.Itoa(child.Type)
	formData["preview"] = "0"
	formData["productType"] = "study"
	formData["scene-type-id"] = fmt.Sprintf("front-3-%d", child.Id)
	formData["termId"] = strconv.FormatInt(chapter.Id, 10)

	result, err := HttpPost("https://course.study.163.com/p/cp/lecture/front/getLectureResource.json", header, func(r *http.Request) {
		AddCookies(r, cookies)
	}, formData)

	if err != nil {
		return
	}

	glog.Info(string(result))

	//获取视频清晰度, 里面会有m3u8不同清晰度的地址 GET
	//"https://vod.study.163.com/eds/api/v1/vod/video?videoId=1214987821&signature=6e6d424f4e2b475a637465574f4961306f4c596f474a416f45444650463674454a5a49396e4d75507a3753564f4371704f346b32454c773768656f474547483268766e5a776a624961344342622f346a5746373662476d47474d52422b5a6d316a4956537772626e32664466575a367468555747347a59544a4f456f2f4558663451463376624e34586d77454a706b6b544a33687838516675636845737351692f744a346b6445534d52553d&clientType=1"
	url := fmt.Sprintf("https://vod.study.163.com/eds/api/v1/vod/video?videoId=%d&signature=%s&clientType=%d", child.Data.ContentId, "", 1)
	videoClarity, err := HttpGet(url, DefaultHeaders, func(r *http.Request) {
		AddCookies(r, cookies)
	})
	if err != nil {
		glog.Error(err)
		return
	}

	glog.Info("视频清晰度 ", string(videoClarity))

	//请求m3u8文件  GET
	//"https://jdvodluwytr3t.vod.126.net/jdvodluwytr3t/nos/ept/hls/2019/08/19/1214987821_cbcee5fc7779461b89e4438772432a51_eshd.m3u8?ak=7909bff134372bffca53cdc2c17adc27a4c38c6336120510aea1ae1790819de816f8b22fb2a630f5a9abc99ea0a26c1e56bac650a4922d83351e84b232884bf29da33a2fceb3a0d16b18820211b50feb04eedd32d103bfabab5d3a1999d53979c7ac434e1ab4f73d44ab4c497dbfe9fee6fc2ae2e15a00d5fd833ea365cf99f4f6c8f6721206d89b7e93d9c01fcec819&token=https%3A%2F%2Fvod.study.163.com%2Feds%2Fapi%2Fv1%2Fvod%2Fhls%2Fkey%3Fid%3D1214987821%26token%3D6ac4053d2860b0203f5b620870516b0acd41bac795822906c1e3c04b829a742140f715ba01304ca42e66b1779f9969e9f75214b0e65b3da892000a0db90c953bf407db0a04221a0a5b475e5c9df80311532c3f3c0e2e635e4a10a532daee0b18671e1bf5f8aa9769ad15bef447b747e37a42404d535868f94a5b4533c2e69801&t=1663767438599"
	url = fmt.Sprintf("%s?ak=%s&t=%d", "", "", time.Now().Unix())
	m3u8, err := HttpGet(url, DefaultHeaders, func(r *http.Request) {
		AddCookies(r, cookies)
	})
	if err != nil {
		glog.Error(err)
		return
	}

	glog.Info("m3u8文件内容 ", string(m3u8))
}
