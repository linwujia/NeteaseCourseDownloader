package main

type M3U8Downloader struct {
	Name      string
	Url       string
	Directory string
}

func NewM3U8Downloader(name, url, directory string) *M3U8Downloader {
	downloader := new(M3U8Downloader)
	return downloader
}
