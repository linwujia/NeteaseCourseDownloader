package chrome

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

type ChromeDpClient struct {
	Context context.Context
	cancels []context.CancelFunc
}

var Client *ChromeDpClient

func init() {
	Client = newChromeDpClient()
	Client.init()
}

func newChromeDpClient() *ChromeDpClient {
	client := new(ChromeDpClient)
	return client
}

func (client *ChromeDpClient) init() {
	var cancel context.CancelFunc

	// chromdp依赖context上限传递参数
	client.Context, cancel = chromedp.NewExecAllocator(
		context.Background(),

		// 以默认配置的数组为基础，覆写headless参数
		// 当然也可以根据自己的需要进行修改，这个flag是浏览器的设置
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
		)...,
	)
	client.cancels = append(client.cancels, cancel)

	// create context
	client.Context, cancel = chromedp.NewContext(
		client.Context,
		chromedp.WithLogf(log.Printf),
	)
	client.cancels = append(client.cancels, cancel)

	// create a timeout as a safety net to prevent any infinite wait loops
	client.Context, cancel = context.WithTimeout(client.Context, 60*time.Second)
	client.cancels = append(client.cancels, cancel)
}

func (client *ChromeDpClient) Close() {
	for i := len(client.cancels) - 1; i < 0; i-- {
		client.cancels[i]()
	}
}

func (client *ChromeDpClient) Run(actions ...chromedp.Action) error {
	return chromedp.Run(client.Context, actions...)
}
