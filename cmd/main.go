package main

import (
	"fmt"
	"gor"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func main() {
	upStr, ok := os.LookupEnv("gor_upstream")
	if !ok {
		log.Fatalln("miss gor_upstream")
	}
	portStr, ok := os.LookupEnv("gor_port")
	var port int
	if ok {
		envPort, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("can't parse gor_port %s, use default port 8080", portStr)
			port = 8080
		} else {
			port = envPort
		}
	} else {
		port = 8080
	}
	upstream, err := url.Parse(upStr)
	if err != nil || (upstream.Scheme != "https" && upstream.Scheme != "http") {
		log.Fatalf("gor_upstream is wrong: %s\n", upStr)
	}
	upConf := gor.NewUpstreamConf(upstream)
	log.Printf("Startup with %s\n", upConf)

	upConf.RunInBackground()

	http.HandleFunc("/", upConf.RProxy)
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		panic(err)
	}
}
