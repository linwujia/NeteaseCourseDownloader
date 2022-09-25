package main

import (
	"flag"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	logDir := flag.String("l", "", "Log directory")
	courseUrl := flag.String("u", "", "Course url")
	*courseUrl = "https://course.study.163.com/480000005355162/learning"
	flag.Parse()

	if *logDir == "" || *logDir == "stderr" {
		flag.Lookup("logtostderr").Value.Set("true")
	} else {
		flag.Lookup("log_dir").Value.Set(*logDir)
	}

	glog.V(3).Info("download course url ", courseUrl)

	manager := NewCourseManager(*courseUrl)
	// 退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		glog.Info("exiting...")
		manager.Stop()
	}()

	manager.Init()
	// 运行课程解析下载
	manager.Run()
}
