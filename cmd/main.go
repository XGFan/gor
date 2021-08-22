package main

import (
	"gor"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	upStr, ok := os.LookupEnv("gor_upstream")
	if !ok {
		log.Fatalln("miss gor_upstream")
	}
	upstream, err := url.Parse(upStr)
	if err != nil || (upstream.Scheme != "https" && upstream.Scheme != "http") {
		log.Fatalf("gor_upstream is wrong: %s\n", upStr)
	}
	upConf := gor.NewUpstreamConf(upstream)
	log.Printf("Startup with %s\n", upConf)

	upConf.RunInBackground()

	http.HandleFunc("/", upConf.RProxy)
	err = http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		panic(err)
	}
}
