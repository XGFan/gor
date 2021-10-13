package gor

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type UpstreamConf struct {
	IPChan chan string

	scheme string
	domain string

	ip   string
	port string

	host string

	proxy *httputil.ReverseProxy
}

func (u *UpstreamConf) String() string {
	return fmt.Sprintf("Scheme: %s, Domain: %s, IP: %s, Port: %s", u.scheme, u.domain, u.ip, u.port)
}

func (u *UpstreamConf) RProxy(w http.ResponseWriter, r *http.Request) {
	newIP := r.Header.Get("X-NEW-IP")
	if newIP != "" {
		log.Printf("request to change ip: %s\n", newIP)
		u.IPChan <- newIP
	}
	u.proxy.ServeHTTP(w, r)
}

func NewUpstreamConf(url *url.URL) *UpstreamConf {
	var port string
	if url.Port() == "" {
		if url.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	} else {
		port = url.Port()
	}

	domain := url.Hostname()
	ip := nsLookup(domain)
	if ip == "" {
		log.Fatalln("Startup fail, can not resolve domain")
	}

	conf := &UpstreamConf{
		scheme: url.Scheme,
		domain: url.Hostname(),

		IPChan: make(chan string, 1),

		ip:   ip,
		port: port,
	}
	conf.generateHost()
	conf.init()
	return conf
}

func (u *UpstreamConf) init() {
	u.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.scheme
			req.URL.Host = u.host
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*2)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (u *UpstreamConf) generateHost() {
	u.host = fmt.Sprintf("%s:%s", u.ip, u.port)
}

func (u *UpstreamConf) TryUpdate(newIP string) {
	ips := []string{u.ip}
	if newIP != "" && newIP != u.ip {
		ips = append(ips, newIP)
	} else {
		nsIP := nsLookup(u.domain)
		if nsIP != "" && nsIP != u.ip {
			ips = append(ips, nsIP)
		}
	}
	if len(ips) > 1 {
		validIP := checkConnection(ips, u.port)
		if validIP != u.ip {
			log.Printf("Update ip from %s to %s\n", u.ip, validIP)
			u.ip = validIP
			u.generateHost()
		} else {
			log.Println("Current IP is still valid")
		}
	}
}

func (u *UpstreamConf) RunInBackground() {
	go func() {
		tick := time.Tick(time.Second * 15)
		for {
			select {
			case newIP := <-u.IPChan:
				u.TryUpdate(newIP)
			case <-tick:
				u.TryUpdate("")
			}
		}
	}()
}

func nsLookup(domain string) string {
	ns, err := net.LookupIP(domain)
	if err != nil {
		log.Printf("nslookup %s fail, %s", domain, err)
		return ""
	}
	if len(ns) == 0 {
		log.Printf("nslookup%s fail, 0 ip return", domain)
		return ""
	}
	return ns[0].String()
}

func checkConnection(ips []string, port string) string {
	respChan := make(chan string, len(ips))
	for _, v := range ips {
		go func(rIP, rPort string) {
			remoteAddr := fmt.Sprintf("%s:%s", rIP, rPort)
			conn, err := net.DialTimeout("tcp", remoteAddr, time.Millisecond*1500)
			defer conn.Close()
			if err != nil {
				return
			}
			_, err = conn.Write([]byte{0x00})
			if err != nil {
				return
			}
			respChan <- rIP
		}(v, port)
	}
	return <-respChan
}
