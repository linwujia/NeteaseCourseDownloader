package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goQrcode "github.com/skip2/go-qrcode"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	LoginUrl = "https://study.163.com/member/login.htm"
)

type AccountUser struct {
	chromeDpClient *ChromeDpClient
	checkLoginUrl  string
}

func NewAccountUser(chromeDpClient *ChromeDpClient) *AccountUser {
	user := new(AccountUser)
	user.chromeDpClient = chromeDpClient
	user.checkLoginUrl = "https://course.study.163.com/480000005355162/learning"
	return user
}

func (user *AccountUser) LoginIfNeed() error {
	return user.chromeDpClient.Run(user.loginEnsureAction())
}

func (user *AccountUser) RedirectToCoursePage() {
	user.chromeDpClient.Run()
}

func (user *AccountUser) loginEnsureAction() chromedp.Tasks {
	return chromedp.Tasks{
		// 1 加载cookies
		user.loadCookies(),
		// 2 确认登录态
		user.loginIfNeed(),
		// 3 保持cookies
		user.saveCookies(),
	}
}

func (user *AccountUser) loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 如果cookies临时文件不存在则直接跳过
		if _, _err := os.Stat("cookies.tmp"); os.IsNotExist(_err) {
			return
		}

		// 如果存在则读取cookies的数据
		cookiesData, err := ioutil.ReadFile("cookies.tmp")
		if err != nil {
			return
		}

		// 反序列化
		cookiesParams := network.SetCookiesParams{}
		if err = cookiesParams.UnmarshalJSON(cookiesData); err != nil {
			return
		}

		// 设置cookies
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
	}
}

func (user *AccountUser) loginIfNeed() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		chromedp.Navigate(user.checkLoginUrl).Do(ctx) //先跳转到vip课程
		//如果未登录，会被调整到其他页面， url就不是这个vip课程url
		var url string
		if err = chromedp.Evaluate(`window.location.href`, &url).Do(ctx); err != nil {
			return
		}
		//已经登录了
		if strings.Contains(url, user.checkLoginUrl) {
			log.Println("已经使用cookies登陆")
			return
		}

		//未登录，先登录
		chromedp.Navigate(LoginUrl).Do(ctx)

		//会有个弹窗，先关闭
		chromedp.Click("#ux-modal > div.ux-modal_ft > span").Do(ctx)

		//微信登录按钮
		if err = chromedp.WaitVisible("#form_parent > div.ux-login-set-login-set-panel > div > div.login-set-panel-login > div > div.ux-urs-login-third.third-login.f-cb > div > a:nth-child(1) > span").Do(ctx); err != nil {
			return
		}

		time.Sleep(1e9)
		//点击微信登录
		if err = chromedp.Click("#form_parent > div.ux-login-set-login-set-panel > div > div.login-set-panel-login > div > div.ux-urs-login-third.third-login.f-cb > div > a:nth-child(1) > span").Do(ctx); err != nil {
			return
		}

		user.getCode().Do(ctx)
		return
	}
}

// 保存Cookies
func (user *AccountUser) saveCookies() chromedp.ActionFunc {

	return func(ctx context.Context) (err error) {
		// 等待二维码登陆
		if err = chromedp.WaitVisible(`#j-topnav > div > div.m-navrgt.f-fr.f-cb.f-pr.j-navright > div.mn-userinfo.f-fr.f-cb.f-pr > div > a`).Do(ctx); err != nil {
			return
		} //

		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return
		}

		// 2. 序列化
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return
		}

		// 3. 存储到临时文件
		if err = ioutil.WriteFile("cookies.tmp", cookiesData, 0755); err != nil {
			return
		}
		return
	}
}

// 获取二维码的过程
func (user *AccountUser) getCode() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 1. 用于存储图片的字节切片
		var code []byte

		// 2. 截图
		// 注意这里需要注明直接使用ID选择器来获取元素（chromedp.ByID）
		if err = chromedp.Screenshot(`#wx_default_tip`, &code, chromedp.ByID).Do(ctx); err != nil {
			return
		}

		// 3. 把二维码输出到标准输出流
		if err = user.printQRCode(code); err != nil {
			return err
		}
		return
	}
}

// 输出二维码
func (user *AccountUser) printQRCode(code []byte) (err error) {
	// 1. 因为我们的字节流是图像，所以我们需要先解码字节流
	img, _, err := image.Decode(bytes.NewReader(code))
	if err != nil {
		return
	}

	// 2. 然后使用gozxing库解码图片获取二进制位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return
	}

	// 3. 用二进制位图解码获取gozxing的二维码对象
	res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return
	}

	// 4. 用结果来获取go-qrcode对象（注意这里我用了库的别名）
	qr, err := goQrcode.New(res.String(), goQrcode.High)
	if err != nil {
		return
	}

	// 5. 输出到标准输出流
	fmt.Println(qr.ToSmallString(false))
	return
}
