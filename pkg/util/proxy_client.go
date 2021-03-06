package util

import (
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
)

//NewTorClient creates a new proxy.Client wrapped by the tor proxy
func NewTorClient(proxyUrl string) *http.Client {
	tbProxyURL, err := url.Parse(proxyUrl)
	Check(err)

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	Check(err)

	tbTransport := &http.Transport{Dial: tbDialer.Dial}
	return &http.Client{Transport: tbTransport}
}
