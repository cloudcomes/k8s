
//https://play.golang.org/p/x6E0T1Hyfz
package main

import (
"fmt"
"log"
"net"
"net/http"
"time"
)

func main() {
	client, transport := NewClient()
	req, err := http.NewRequest("HEAD", "http://baidu.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	n2 := time.Now()
	resp2, err := transport.RoundTrip(req)
	if err != nil {
		log.Fatal(err)
	}
	resp2.Body.Close()
	scode2 := resp2.StatusCode
	stext2 := http.StatusText(resp2.StatusCode)
	fmt.Println(scode2, stext2)
	fmt.Println("RoundTrip took:", time.Since(n2))

	println()

	n1 := time.Now()
	resp1, err := client.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp1.Body.Close()
	scode1 := resp1.StatusCode
	stext1 := http.StatusText(resp1.StatusCode)
	fmt.Println(scode1, stext1)
	fmt.Println("Client took:", time.Since(n1))
}

/*
301 Moved Permanently
RoundTrip took: 50.905464ms

200 OK
Client took: 124.773167ms
*/

type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient() (*Client, *http.Transport) {
	dialfunc := func(network, addr string) (net.Conn, error) {
		cn, err := net.DialTimeout(network, addr, time.Second*5)
		if err != nil {
			log.Fatal(err)
		}
		return cn, err
	}
	transport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		Dial:              dialfunc,
		DisableKeepAlives: true,
	}
	client := &Client{}
	client.httpClient = &http.Client{
		Transport: transport,
	}
	return client, transport
}

