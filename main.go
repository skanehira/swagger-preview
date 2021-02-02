package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func detectFileType(ext string) string {
	ext = strings.ToLower(ext[1:])
	switch ext {
	case "yaml", "yml":
		return "swagger"
	case "md", "markdown":
		return "markdown"
	case "adoc", "asciidoc":
		return "asciidoc"
	}
	return ext
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "please specify file")
		os.Exit(1)
	}

	fileName := os.Args[1]

	msg := make(chan []byte)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fi, err := os.Stat(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	old := fi.ModTime()

	go func() {
		var once sync.Once
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				go func() {
					once.Do(func() {
						<-time.After(50 * time.Millisecond)
						if event.Name == fileName {
							fi, err := os.Stat(fileName)
							if err != nil {
								log.Println(err)
								return
							}
							now := fi.ModTime()
							if !old.Equal(now) {
								old = now

								fmt.Println("update...")
								b, err := ioutil.ReadFile(fileName)
								if err != nil {
									log.Println(err)
									return
								}
								msg <- b
							}
						}
					})
					once = sync.Once{}
				}()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(filepath.Dir(fileName))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		// send file name at first
		if err := c.WriteJSON(map[string]string{"fileName": fileName, "fileType": detectFileType(filepath.Ext(fileName))}); err != nil {
			log.Println(err)
			return
		}

		for {
			msg := <-msg
			if err := c.WriteJSON(map[string]string{"message": string(msg)}); err != nil {
				log.Println(err)
				return
			}
		}
	})
	http.Handle("/", http.FileServer(http.Dir(".")))
	log.Println("start http server :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
