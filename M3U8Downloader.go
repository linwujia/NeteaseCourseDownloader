package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type M3U8Downloader struct {
	Name      string
	Url       string
	Host      string
	Key       []byte
	Content   string
	Directory string
	Cookies   []*http.Cookie
}

func NewM3U8Downloader(name, url string, key []byte, m3u8Content, directory string, cookies []*http.Cookie) *M3U8Downloader {
	downloader := new(M3U8Downloader)
	downloader.Name = name
	downloader.Url = url
	downloader.Host = getHost(url)
	downloader.Key = getKey( /*[]byte("DuzBzHF5RcOHVgoRWV0hNA==")*/ key)
	downloader.Content = m3u8Content
	downloader.Directory = directory
	downloader.Cookies = cookies
	/*glog.Info(hex.DecodeString(string(key)))*/
	return downloader
}

func getKey(key []byte) []byte {
	/*l := len(key)
	mod := l % 4
	result := make([]byte, l + (4 - mod))
	copy(result[:l], key)

	for i := 0; i < 4 - mod; i++ {
		result[l + i] = byte(0)
	}*/
	decode := make([]byte, base64.StdEncoding.DecodedLen(len(key)))
	n, err := base64.StdEncoding.Decode(decode, key)
	if err != nil {
		return nil
	}

	return decode[:n]
}

func getHost(u string) string {
	url, err := url.Parse(u)
	if err != nil {
		return ""
	}

	path := url.Path
	index := strings.LastIndex(path, "/")
	path = path[:index]
	return fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, path)
}

func (downloader *M3U8Downloader) Download() {
	tsList := downloader.Prepare()
	glog.Info("待下载 ts 文件数量:", len(tsList))

	// 下载 ts 文件
	var wg sync.WaitGroup
	//max := make(chan bool)
	for _, info := range tsList {
		wg.Add(1)
		//max <- true
		go func(tsInfo TsInfo) {
			defer func() {
				wg.Done()
				//<- max
			}()
			downloader.downloadTsFile(tsInfo)
		}(info)
	}
	wg.Wait()

	glog.Info("all ts file download completed")
	// 合并 ts切割文件成mp4文件
}

func (downloader *M3U8Downloader) Prepare() (tsList []TsInfo) {
	return downloader.getTsList()
}

func (downloader *M3U8Downloader) getTsList() (tsList []TsInfo) {
	lines := strings.Split(downloader.Content, "\n")
	var ts TsInfo
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			if strings.HasPrefix(line, "http") || strings.HasPrefix(line, "https") {
				ts = TsInfo{
					Name: line,
					Url:  line,
				}
				tsList = append(tsList, ts)
			} else {
				ts = TsInfo{
					Name: line,
					Url:  fmt.Sprintf("%s/%s", downloader.Host, line),
				}
				tsList = append(tsList, ts)
			}
		}
	}

	return
}

func (downloader *M3U8Downloader) downloadTsFile(info TsInfo) {
	glog.Info("start download ts file ", info.Name)
	orignData, err := HttpGet(info.Url, DefaultHeaders, downloader.AddCookies)
	if err != nil {
		glog.Error(err)
		return
	}

	orignData, err = AesDecrypt(orignData, downloader.Key)
	if err != nil {
		glog.Error(err)
		return
	}

	syncByte := uint8(71) //0x47
	bLen := len(orignData)
	for j := 0; j < bLen; j++ {
		if orignData[j] == syncByte {
			orignData = orignData[j:]
			break
		}
	}
	ioutil.WriteFile(info.Name, orignData, 0666)
	glog.Info("download ts file ", info.Name, " completed")
}

func (downloader *M3U8Downloader) AddCookies(r *http.Request) {
	AddCookies(r, downloader.Cookies)
}

func AesDecrypt(crypted, key []byte, ivs ...[]byte) ([]byte, error) {
	if len(key) > 16 {
		key = key[:16]
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	var iv []byte
	if len(ivs) == 0 {
		iv = key
	} else {
		iv = ivs[0]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

type TsInfo struct {
	Name string
	Url  string
}
