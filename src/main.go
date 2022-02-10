package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var domain string
var translatePath string

var agentResultTime = make(map[string]int64)
var agentResult = make(map[string][]byte)
var agentHeader = make(map[string]http.Header)

// 网页缓存时间 秒
var agentCacheTime int64 = 24 * 60 * 60

var fileJsonMapTime = make(map[string]int64)
var fileJsonMap = make(map[string]map[string]string)

// 文件缓存时间 秒
var fileCacheTime int64 = 5 * 60 * 60

func main() {

	flagConfig()

	http.HandleFunc("/", transit(domain, translatePath))
	log.Fatal(http.ListenAndServe(":8086", nil))

}

func flagConfig() {
	flag.StringVar(&domain, "domain", "https://docs.rainmeter.net", "--domain http://xxx.xxx.xxx")
	flag.StringVar(&translatePath, "translate", "./translate", "--translate /data/translate")
	flag.Parse()
}

func transit(domain, translatePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url := domain + r.URL.Path
		t := time.Now().Unix()
		if agentResultTime[url] == 0 || t-agentResultTime[url] > agentCacheTime {
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			agentResult[url] = body
			agentHeader[url] = resp.Header
			agentResultTime[url] = t
		}
		if r.URL.Path == "/img/logo_nav.png" {
			w.Write(logo(translatePath))
			return
		}
		if agentHeader[url] != nil {
			for k, v := range agentHeader[url] {
				w.Header().Set(k, v[0])
			}
		}

		rbs := translate(agentResult[url], translatePath+"/common.json")
		rbs = translate(rbs, translatePath+r.URL.Path+"/file.json")
		w.Write(rbs)
	}
}

func translate(result []byte, translatePath string) []byte {

	var jsonMap map[string]string
	t := time.Now().Unix()

	if fileJsonMapTime[translatePath] == 0 || t-fileJsonMapTime[translatePath] > fileCacheTime {
		file, err := os.Open(translatePath)
		if err != nil {
			return result
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		jsonString := string(content)

		json.Unmarshal([]byte(jsonString), &jsonMap)
		fileJsonMap[translatePath] = jsonMap
		fileJsonMapTime[translatePath] = t
	} else {
		jsonMap = fileJsonMap[translatePath]
	}

	resultStr := string(result)
	for k, v := range jsonMap {
		resultStr = strings.ReplaceAll(resultStr, k, v)
	}
	return []byte(resultStr)
}

func logo(translatePath string) []byte {
	file, err := os.Open(translatePath + "/logo_nav.png")
	if err != nil {
		return nil
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	return content
}
