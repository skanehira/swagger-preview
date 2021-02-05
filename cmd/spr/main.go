package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

func OpenBrowser(url string) error {
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		r := strings.NewReplacer("&", "^&")
		args = []string{"cmd", "start", "/", r.Replace(url)}
	case "linux":
		args = []string{"xdg-open", url}
	case "darwin":
		args = []string{"open", url}
	}

	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

var upgrader = websocket.Upgrader{}

var indexHTML = `
<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <title>mpr</title>
</head>
<script src="https://cdn.jsdelivr.net/npm/vue@2.6.12"></script>
<link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.41.1/swagger-ui.css" >
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.41.1/swagger-ui-bundle.js"> </script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.41.1/swagger-ui-standalone-preset.js"> </script>
<body>
  <div id="app">
    <div id="ui"></div>
  </div>
</body>
<script>
const app = new Vue({
  el: '#app',
  data: {
    ui: {},
    ws: {},
  },
  methods: {
  },
  mounted(){
    this.ws = new WebSocket("ws://localhost:%s/ws")

    let isFirst = false
    this.ws.onmessage = (ev) => {
      const resp = JSON.parse(ev.data)
      if (!isFirst) {
        this.ui = SwaggerUIBundle({
          url : resp.fileName,
          dom_id: '#ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        })
        isFirst = true
        return
      }

      console.log("update");
      this.ui.specActions.updateSpec(resp.message)
    }

    this.ws.onerr = (err) => {
      console.log(err)
    }

    window.onbeforeunload = () => {
      this.ws.send(0)
    }
  }
})
</script>
</html>
`

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
						<-time.After(100 * time.Millisecond)
						if filepath.Base(event.Name) == filepath.Base(fileName) {
							fi, err := os.Stat(fileName)
							if err != nil {
								log.Println(err)
								return
							}
							now := fi.ModTime()
							if !old.Equal(now) {
								old = now

								log.Println("update")
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

		resp := map[string]interface{}{
			"fileName": fileName,
		}
		if err := c.WriteJSON(resp); err != nil {
			log.Println(err)
			return
		}

		done := make(chan bool)
		go func() {
			// close websocket when recive some message
			c.ReadMessage()
			done <- true
		}()
		for {
			select {
			case msg := <-msg:
				if err := c.WriteJSON(map[string]string{"message": string(msg)}); err != nil {
					log.Println(err)
					return
				}
			case <-done:
				log.Println("close websocket")
				return
			}
		}
	})

	port := "9999"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(indexHTML, port)
		w.Write([]byte(body))
		return
	})
	log.Println("start server:", port)
	log.Println("watching", fileName)

	if err := OpenBrowser("http://localhost:" + port); err != nil {
		log.Println("cannot open browser", err)
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
