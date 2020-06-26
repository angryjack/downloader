package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	streams  int    = 30                     // max simultaneous streams
	file     string = "files.txt"            // file with links
	path     string = "/Users/me/Downloads/" // path to save
	attempts int    = 10                     // max attempts to download
)

func main() {
	downloadFiles()
}

func downloadFiles() {
	bytes, _ := ioutil.ReadFile(file)
	files := strings.Split(string(bytes), "\n")

	var tokens = make(chan struct{}, streams)
	var ch = make(chan bool)
	count := len(files)

	i := 1
	for _, f := range files {
		fmt.Printf("\rDownloading: [%d:%d]", i, count)
		tokens <- struct{}{}
		go downloadFile(f, tokens, 1)
		i++
	}

	for range ch {
		if i > count {
			close(ch)
		}
	}
}

func downloadFile(f string, tokens chan struct{}, t int) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("fail:", err)
		}
	}()
	if t > attempts {
		fmt.Println("max attempts reached!")
		return
	}
	resp, err := httpClient().Get(f)
	checkError(err)
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		file := createFile(path + prepareFileName(f))
		defer file.Close()
		_, err = io.Copy(file, resp.Body)
		checkError(err)
	} else {
		time.Sleep(1 * time.Second)
		downloadFile(f, tokens, t+1)
	}

	<-tokens
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func createFile(fileName string) *os.File {
	f, err := os.Create(fileName)

	checkError(err)
	return f
}

func prepareFileName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}
