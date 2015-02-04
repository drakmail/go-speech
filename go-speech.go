package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var downloadingFiles map[string]bool

func fileSize(file string) int64 {
	f, err := os.Open(file)
	if err != nil {
		log.Print("Couldn't open file")
		return 0
	}
	fi, err := f.Stat()
	if err != nil {
		log.Print("Couldn't stat file")
	}
	return fi.Size()
}

func callback(filename, callbackURL string) {
	_, err := http.PostForm(callbackURL, url.Values{"filename": {filename}})
	if err != nil {
		log.Print("Couldn't send callback to " + callbackURL)
	}
}

// TODO: Refactor :)
func downloadFile(fileURL, filename, callbackURL string, retries int32) bool {
	if retries >= 0 {
		time.Sleep(1000 * time.Millisecond)
		if !downloadingFiles[filename] {
			log.Print("started downloading file to ", filename)
			downloadingFiles[filename] = true
			resp, err := http.Get(fileURL)
			if err != nil {
				log.Print("Couldn't fetch url: ", err)
				downloadingFiles[filename] = false
				return downloadFile(fileURL, filename, callbackURL, retries-1)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print("Couldn't read body: ", err)
				downloadingFiles[filename] = false
				return downloadFile(fileURL, filename, callbackURL, retries-1)
			}
			err = ioutil.WriteFile(filename, body, 0777)
			if err != nil {
				log.Print("Couldn't create and write to a cache file: ", err)
				downloadingFiles[filename] = false
				return downloadFile(fileURL, filename, callbackURL, retries-1)
			}
			log.Print("Downloading file complete ", filename)
			downloadingFiles[filename] = false
			// TODO: if downloaded file is valid mp3 with length > 1 sec - save MD5.mp3 and serve it
			// if downloaded file is > 4000 bytes length then all ok
			if fileSize(filename) > 4000 {
				callback(filename, callbackURL)
				return true
			} else {
				// redownload it
				return downloadFile(fileURL, filename, callbackURL, retries-1)
			}
		} else {
			log.Print("File already downloading...")
			return false
		}
	} else {
		log.Print("Couldn't download file after 3 retries =(")
		return false
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
		// if need to send callback - send it immedately
		if requestCallback != "" {
			go callback(savedFilename, requestCallback)
		}
	} else {
		// else - try to download
		log.Print("Downloading result of ", proxyRequest)
		if requestCallback != "" {
			go downloadFile(proxyRequest, savedFilename, requestCallback, 3)
		} else {
			go downloadFile(proxyRequest, savedFilename, "https://posthere.io/d53c-48ad-90db", 3)
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
	log.Fatal(http.ListenAndServe("0.0.0.0:8989", nil))
}
