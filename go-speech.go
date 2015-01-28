package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var downloadingFiles map[string]bool

func callback(filename, callbackURL string) {
	_, err := http.PostForm(callbackURL, url.Values{"filename": {filename}})
	if err != nil {
		log.Print("Couldn't send callback to " + callbackURL)
	}
}

func downloadFile(fileURL, filename, callbackURL string) {
	if !downloadingFiles[filename] {
		log.Print("started downloading file ", fileURL)
		downloadingFiles[filename] = true
		resp, err := http.Get(fileURL)
		if err != nil {
			log.Print("Couldn't fetch url: ", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print("Couldn't read body: ", err)
		}
		err = ioutil.WriteFile(filename, body, 0777)
		if err != nil {
			log.Print("Couldn't create and write to a temp file: ", err)
		}
		log.Print("Downloading file complete", fileURL)
		downloadingFiles[filename] = false
		// if downloaded file is valid mp3 with length > 1 sec - save MD5.mp3 and serve it
		callback(filename, callbackURL)
	} else {
		log.Print("File already downloading...")
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var fileExists bool
	var requestCallback string
	// typical request: https://tts.voicetech.yandex.net/generate?speaker=zahar&key=API_KEY&format=mp3&lang=ru-RU&text=SOMETEXT
	requestString := r.URL.RawQuery
	r.ParseForm()
	if len(r.Form["callback"]) > 0 {
		requestCallback = r.Form["callback"][0]
	} else {
		requestCallback = ""
	}
	proxyRequest := "https://tts.voicetech.yandex.net/generate?" + requestString
	// when get request - calculate it md5
	requestHash := md5.Sum([]byte(requestString))
	savedFilename := fmt.Sprintf("cache/%x.mp3", requestHash)
	// then try to find file MD5.mp3 in cache
	if _, err := os.Stat(savedFilename); os.IsNotExist(err) {
		fileExists = false
	} else {
		fileExists = true
	}
	if fileExists {
		// if found - serve it
		log.Print("Serving file")
		http.ServeFile(w, r, savedFilename)
	} else {
		// else - try to download
		log.Print("Downloading result of ", proxyRequest)
		if requestCallback != "" {
			go downloadFile(proxyRequest, savedFilename, requestCallback)
		} else {
			go downloadFile(proxyRequest, savedFilename, "https://posthere.io/d53c-48ad-90db")
		}
	}
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/cache/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Request for file ", r.URL.Path[1:])
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	// listen on port
	err := os.MkdirAll("cache", 0777)
	if err != nil {
		log.Print("cache directory creation failed: ", err)
	}
	log.Print("started on http://0.0.0.0:8989")
	downloadingFiles = map[string]bool{}
	log.Fatal(http.ListenAndServe(":8989", nil))
}
