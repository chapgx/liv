package wserv

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chapgx/liv/pkg/wserv/web"
	"github.com/chapgx/owl"
	"github.com/gorilla/websocket"
)

const (
	PORT = "9890"
	Host = "127.0.0.1"
)

var Done = make(chan int, 1)
var basedir string
var indexPage string
var sub owl.Subscriber
var currentpage string

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// allow all for development
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, e := upgrader.Upgrade(w, r, nil)
	if e != nil {
		log.Println("Upgrader error", e)
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	// read pump
	go func() {
		defer close(done)
		for {
			if _, _, e := conn.ReadMessage(); e != nil {
				return
			}
		}
	}()

	//TODO: implement better logging
	fmt.Println("connection open")
	for {
		select {
		case rslt, isOpen := <-sub.Listen():
			if !isOpen {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				fmt.Println("subscription ended")
				return
			}
			switch d := rslt.(type) {
			case error:
				continue
			case owl.ReadSnap:
				msg := []byte("refresh:")
				if filepath.Base(d.Path) == currentpage {
					msg = []byte("update:")
					msg = append(msg, d.Content...)
				}

				if e := conn.WriteMessage(websocket.TextMessage, msg); e != nil {
					fmt.Println("error sending message", e)
					return
				}
				fmt.Println("signal sent to client")
			}
		case <-done:
			fmt.Println("client closed connection")
			return
		}
	}
}

func RunServer(dir, rootfile string) {
	mx := http.NewServeMux()

	mx.HandleFunc("/ws", handleConnections)
	handlerFileChange := changeFile(mx)
	handler := serverFile(handlerFileChange)

	server := http.Server{Addr: Host + ":" + PORT, Handler: handler}

	fmt.Println("server running on", PORT)
	basedir = dir
	indexPage = rootfile

	if indexPage == "" {
		indexPage = filepath.Base(basedir)
	}

	sub = owl.SubscribeOnModified(owl.R_READ)
	go owl.WatchWithMinInterval(dir)

	if e := server.ListenAndServe(); e != nil {
		fmt.Println(e)
		Done <- 1
	}
}

func changeFile(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}

		b, e := os.ReadFile(filepath.Join(basedir, r.URL.Path))
		if e != nil {
			fmt.Println("form web server", e)
			Done <- 1
		}

		bts, e := web.WWW.ReadFile("www/index.html")
		if e != nil {
			panic(e)
		}
		bodystart := bytes.Index(b, []byte("<body>"))
		bodyend := bytes.Index(b, []byte("</body>"))

		//BUG: body with class fails
		if bodystart == -1 || bodyend == -1 {
			panic("malformed file")
		}
		bodystart += 6

		headstart := bytes.Index(b, []byte("<head>"))
		headend := bytes.Index(b, []byte("</head>"))

		if headstart == -1 || headend == -1 {
			panic("malformed file")
		}
		headstart += 6

		strbody := string(bts)
		strbody = strings.Replace(strbody, "{head}", string(b[headstart:headend]), 1)
		strbody = strings.Replace(strbody, "{body}", string(b[bodystart:bodyend]), 1)

		currentpage = strings.TrimPrefix(r.URL.Path, "/")

		w.Write([]byte(strbody))
	})
}

func serverFile(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			r.URL.Path = "/" + indexPage
		}

		path := filepath.Join(basedir, strings.TrimPrefix(r.URL.Path, "/"))

		ext := filepath.Ext(path)
		if ext == "" || ext == ".html" || ext == "ws" {
			next.ServeHTTP(w, r)
			return
		}

		f, e := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if e != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		stat, _ := f.Stat()

		http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)
	})
}
