package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	chiprometheus "github.com/tliefheid/prometheus-chi-metric-middleware"
)

func main() {
	r := chi.NewRouter()

	// create the middleware
	metricMiddleware := chiprometheus.NewMiddleware(chiprometheus.Options{
		Namespace: "mock",
	})

	// use the middlware
	r.Use(metricMiddleware.Handler)

	// show metrics on /metrics
	r.Handle("/metrics", promhttp.Handler())

	// demo endpoints
	r.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		randomSleep()
		w.Write([]byte("foo"))
	})

	r.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		randomSleep()
		w.Write([]byte("/user/{id}"))
	})

	r.Get("/404", func(w http.ResponseWriter, r *http.Request) {
		randomSleep()
		w.WriteHeader(http.StatusNotFound)
	})
	f := NewFoo()
	r.Get("/*", f.handler)

	fmt.Println("start webserver on port 8080")
	http.ListenAndServe("127.0.0.1:8080", r)
}

func randomSleep() {
	sleep := rand.Intn(200) + 1
	time.Sleep(time.Duration(sleep) * time.Millisecond)
}

type foo struct {
	bar string
}

func NewFoo() *foo {
	f := new(foo)
	f.bar = "foobar"
	return f
}

func (f *foo) handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())
	w.Write([]byte("done"))
}
