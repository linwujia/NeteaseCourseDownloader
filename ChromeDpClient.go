package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

type ChromeDpClient struct {
	context context.Context
	cancels []context.CancelFunc
}

func NewChromeDpClient() *ChromeDpClient {
	client := new(ChromeDpClient)
	return client
}

func (client *ChromeDpClient) Init() {
	var cancel context.CancelFunc

	// chromdp依赖context上限传递参数
	client.context, cancel = chromedp.NewExecAllocator(
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
	client.context, cancel = chromedp.NewContext(
		client.context,
		chromedp.WithLogf(log.Printf),
	)
	client.cancels = append(client.cancels, cancel)

	// create a timeout as a safety net to prevent any infinite wait loops
	client.context, cancel = context.WithTimeout(client.context, 60*time.Second)
	client.cancels = append(client.cancels, cancel)
}

func (client *ChromeDpClient) Close() {
	for i := len(client.cancels) - 1; i < 0; i-- {
		client.cancels[i]()
	}
}

func (client *ChromeDpClient) Run(actions ...chromedp.Action) error {
	return chromedp.Run(client.context, actions...)
}
