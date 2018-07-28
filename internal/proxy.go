package internal

import (
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
)

func NewProxyClient() *http.Client {
	tbProxyURL, err := url.Parse("socks5://127.0.0.1:9050")
	Check(err)

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	Check(err)

	tbTransport := &http.Transport{Dial: tbDialer.Dial}
	return &http.Client{Transport: tbTransport}
}
