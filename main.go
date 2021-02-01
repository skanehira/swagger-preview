package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	fi, err := os.Stat(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	old := fi.ModTime()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			<-ticker.C
			fi, err := os.Stat(fileName)
			if err != nil {
				log.Println(err)
				continue
			}
			now := fi.ModTime()
			if !old.Equal(now) {
				old = now

				fmt.Println("update...")
				b, err := ioutil.ReadFile(fileName)
				if err != nil {
					log.Println(err)
					continue
				}
				msg <- b
			}
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		// send file name at first
		c.ReadMessage()
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
