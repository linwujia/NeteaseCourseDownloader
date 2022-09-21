package main

import (
	chrome2 "NeteaseCourseDownloader/chrome"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WebDriver struct {
	service *selenium.Service
	selenium.WebDriver
}

func Init(path string, port int) (chrome2.IWebDriver, error) {
	//如果seleniumServer没有启动，就启动一个seleniumServer所需要的参数，可以为空，示例请参见https://github.com/tebeka/selenium/blob/master/example_test.go
	var opts []selenium.ServiceOption
	//opts := []selenium.ServiceOption{
	//    selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
	//    selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
	//}

	//selenium.SetDebug(true)
	service, err := selenium.NewChromeDriverService(path, port, opts...)
	if nil != err {
		fmt.Println("start a chromedriver service falid", err.Error())
		return nil, err
	}

	driver := new(WebDriver)
	driver.service = service
	driver.WebDriver, err = selenium.NewRemote(*initCaps(), fmt.Sprintf("http://localhost:%d/wd/hub", port))

	return driver, err
}

func (d *WebDriver) Quit() error {
	defer d.service.Stop()
	if d.WebDriver != nil {
		defer d.WebDriver.Quit()
	}
	return nil
}

func (d *WebDriver) Open(url string) (err error) {
	err = d.Get(url)
	return
}

func (d *WebDriver) ClickXPath(path string) {
	err := d.Wait(func(wd selenium.WebDriver) (bool, error) {
		element, err := wd.FindElement(selenium.ByXPATH, path)
		if err != nil {
			return false, err
		}

		return element.IsDisplayed()
	})

	if err != nil {
		glog.Error("click x path 出了问题", err)
		return
	}

	element, err := d.FindElement(selenium.ByXPATH, path)
	if err != nil {
		glog.Error("click x path 出了问题", err)
		return
	}
	element.Click()
}

func (d *WebDriver) GetXPathText(path string) (text string, err error) {
	err = d.Wait(func(wd selenium.WebDriver) (bool, error) {
		element, err := wd.FindElement(selenium.ByXPATH, path)
		if err != nil {
			return false, err
		}

		return element.IsDisplayed()
	})

	if err != nil {
		glog.Error("get x path text 出了问题", err)
		return
	}

	element, err := d.FindElement(selenium.ByXPATH, path)
	if err != nil {
		glog.Error("get x path text 出了问题", err)
		return
	}

	if text, err = element.Text(); err != nil {
		glog.Error("get x path text 出了问题", err)
		return
	}

	return
}

func (d *WebDriver) ViewTips() ([]string, string) {
	// global answer
	tipsOpen, err := d.FindElement(selenium.ByXPATH, `//*[@id="app"]/div/div[*]/div/div[*]/div[*]/div[*]/span[contains(text(), "查看提示")]`)
	if err != nil {
		glog.Error("没有可点击的【查看提示】按钮", err)
		return d.tryGetAnswers()
	}

	tipsOpen.Click()
	glog.Info("有可点击的【查看提示】按钮")

	time.Sleep(time.Second)

	tipsOpen, err = d.FindElement(selenium.ByXPATH, `//*[@id="app"]/div/div[*]/div/div[*]/div[*]/div[*]/span[contains(text(), "查看提示")]`)
	if err != nil {
		glog.Error("关闭查看提示失败！没有可点击的【查看提示】按钮", err)
		return []string{}, ""
	}

	tipsOpen.Click()
	glog.Info("关闭查看提示成功， 有可点击的【查看提示】按钮")
	time.Sleep(time.Second)

	tipDiv, err := d.FindElement(selenium.ByCSSSelector, ".ant-popover .line-feed")
	if err != nil {
		return []string{}, ""
	}
	tipFullText, err := tipDiv.GetAttribute("innerHTML")
	if err != nil {
		return []string{}, ""
	}

	re := regexp.MustCompile(`</font[a-zA-Z]*?><font+.*?>`)
	html := re.ReplaceAllString(tipFullText, "")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return []string{}, ""
	}

	var answers []string
	doc.Find("font").Each(func(i int, selection *goquery.Selection) {
		answers = append(answers, selection.Text())
	})

	glog.Info("获取提示：", answers)
	time.Sleep(time.Second)

	_, err = d.FindElement(selenium.ByCSSSelector, `.ant-popover-hidden`) //关闭tip则为hidden
	if err != nil {                                                       //没有关闭tip
		tipClose, err := d.FindElement(selenium.ByXPATH, `//*[@id="app"]/div/div[2]/div/div[4]/div[1]/div[1]`)
		if err != nil {
			glog.Error("没有可点击的【关闭提示】按钮")
		} else {
			tipClose.Click()
		}
	}

	return answers, tipFullText
}

func (d *WebDriver) tryGetAnswers() ([]string, string) {
	answerElement, err := d.FindElement(selenium.ByCSSSelector, ".answer")
	if err != nil {
		return []string{}, ""
	}

	text, err := answerElement.Text()
	if err != nil {
		return []string{}, ""
	}

	var answers []string
	text = SubString(text, 5, len(text)-5)

	answerList := strings.Split(text, " ")

	ansOptions, err := d.RadioGetOptions()
	for _, option := range ansOptions {
		for _, ans := range answerList {
			if ans == string(option[0]) {
				answers = append(answers, option)
			}
		}

	}
	glog.Info("找到答案解析：", answers)
	return answers, ""
}

func (d *WebDriver) RadioGetOptions() (options []string, err error) {
	html, err := d.PageSource()
	if err != nil {
		return
	}

	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	document.Find("div .choosable").Each(func(i int, selection *goquery.Selection) {
		options = append(options, selection.Text())
	})

	if len(options) <= 0 {
		document.Find("div .q-answer").Each(func(i int, selection *goquery.Selection) {
			options = append(options, selection.Text())
		})
	}
	glog.Info("获取选项：", options)
	return
}

func (d *WebDriver) FillInBlank(answers []string) {
	for i, answer := range answers {
		input, err := d.FindElement(selenium.ByXPATH, "//*[@id=\"app\"]/div/div[2]/div/div[4]/div[1]/div[2]/div/input["+strconv.Itoa(i+1)+"]")
		if err != nil {
			input, err = d.FindElement(selenium.ByXPATH, "//*[@id=\"app\"]/div/div[2]/div/div[4]/div[1]/div[3]/div/input["+strconv.Itoa(i+1)+"]")
			if err != nil {
				continue
			}
		}

		input.SendKeys(answer)
	}

	d.checkDelay()
	var submits []selenium.WebElement
	err := d.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		element, err := wd.FindElement(selenium.ByCSSSelector, ".action-row")
		if err != nil {
			return false, err
		}

		submits, err = element.FindElements(selenium.ByXPATH, "button")
		if err != nil {
			return false, err
		}

		return len(submits) > 0, nil
	}, 15*time.Second)

	if err != nil {
		return
	}

	if len(submits) > 1 {
		d.ClickXPath(`//*[@id="app"]/div/div[2]/div/div[6]/div[2]/button[2]`)
		glog.Info("成功点击交卷！")
	} else {
		d.ClickXPath(`//*[@id="app"]/div/div[*]/div/div[*]/div[*]/button`)
		glog.Info("点击进入下一题")
	}

}

func (d *WebDriver) SpecialFillInBlank(answers []string) {
	for i, answer := range answers {
		input, err := d.FindElement(selenium.ByXPATH, "//*[@id=\"app\"]/div/div[2]/div/div[6]/div[1]/div[2]/div/input["+strconv.Itoa(i+1)+"]")
		if err != nil {
			continue
		}

		input.SendKeys(answer)
	}

	d.checkDelay()
	var submits []selenium.WebElement
	err := d.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		element, err := wd.FindElement(selenium.ByCSSSelector, ".action-row")
		if err != nil {
			return false, err
		}

		submits, err = element.FindElements(selenium.ByXPATH, "button")
		if err != nil {
			return false, err
		}

		return len(submits) > 0, nil
	}, 15*time.Second)

	if err != nil {
		return
	}

	if len(submits) > 1 {
		d.ClickXPath(`//*[@id="app"]/div/div[2]/div/div[6]/div[2]/button[2]`)
		glog.Info("成功点击交卷！")
	} else {
		d.ClickXPath(`//*[@id="app"]/div/div[*]/div/div[*]/div[*]/button`)
		glog.Info("点击进入下一题")
	}
}

func (d *WebDriver) checkDelay() {
	rand.Seed(time.Now().Unix())
	delay := rand.Intn(3) + 2
	time.Sleep(time.Duration(delay) * time.Second)
}

func (d *WebDriver) RadioCheck(radioTips string) {
	opts, err := d.FindElements(selenium.ByCSSSelector, ".choosable")
	if err != nil {
		return
	}

	for _, b := range []byte(radioTips) {
		for _, opt := range opts {
			text, err := opt.Text()
			if err != nil {
				glog.Error("点击", string(b), "失败！")
				break
			}
			if strings.Contains(text, string(b)) {
				err = opt.Click()
				if err != nil {
					glog.Error("点击", string(b), "失败！")
					break
				}
			}
		}
	}

	d.checkDelay()

	var submits []selenium.WebElement
	err = d.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		element, err := wd.FindElement(selenium.ByCSSSelector, ".action-row")
		if err != nil {
			return false, err
		}

		submits, err = element.FindElements(selenium.ByXPATH, "button")
		if err != nil {
			return false, err
		}

		return len(submits) > 0, nil
	}, 15*time.Second)

	if err != nil {
		return
	}

	if len(submits) > 1 {
		d.ClickXPath(`//*[@id="app"]/div/div[2]/div/div[6]/div[2]/button[2]`)
		glog.Info("成功点击交卷！")
	} else {
		d.ClickXPath(`//*[@id="app"]/div/div[*]/div/div[*]/div[*]/button`)
		glog.Info("点击进入下一题")
	}

	time.Sleep(time.Second)
	if elements, err := d.FindElements(selenium.ByCSSSelector, "nc-mask-display"); err == nil && len(elements) > 0 {
		glog.Fatalf("出现滑块验证，本次答题结束")
	}
}

func initCaps() *selenium.Capabilities {
	//链接本地的浏览器 chrome
	caps := &selenium.Capabilities{
		"browserName": "chrome",
	}

	//禁止图片加载，加快渲染速度 2 加载图片， 1 禁止图片加载
	imagCaps := map[string]interface{}{
		"profile.managed_default_content_settings.images": 1,
	}
	chromeCaps := chrome.Capabilities{
		Prefs: imagCaps,
		Path:  "",
		Args: []string{
			//"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			//"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36", // 模拟user-agent，防反爬
			"--excludeSwitches=enable-automation",
		},
	}
	//以上是设置浏览器参数
	caps.AddChrome(chromeCaps)

	return caps
}

func SubString(str string, begin, length int) string {
	glog.Info("Substring =", str)
	rs := []rune(str)
	lth := len(rs)
	glog.Infof("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}
	glog.Infof("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	return string(rs[begin:end])
}
