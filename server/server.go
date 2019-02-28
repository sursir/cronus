package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Middleware struct
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Logging logs all request
func Logging() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestStartedAt := time.Now()
			defer func() {
				log.Println(r.URL.Path, time.Since(requestStartedAt))
			}()

			f(w, r)
		}
	}
}

// Authenticate verify request
func Authenticate(token string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestToken := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "Bearer"))
			if requestToken != token {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			f(w, r)
		}
	}
}

// Chain 执行 Middleware 和 HandleFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}

	return f
}

func main() {
	listenPort := flag.String("port", "8080", "Http listen port")
	token := flag.String("token", "", "验证 Token")

	flag.Parse()

	http.HandleFunc("/", Chain(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		fmt.Println(string(body))
	}, Authenticate(*token), Logging()))

	log.Println(http.ListenAndServe(":"+strings.TrimLeft(*listenPort, ":"), nil))
}
