package http_proxy

import (
	"github.com/demeyerthom/belgian-companies/pkg/errors"
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
)

func NewProxyClient() *http.Client {
	tbProxyURL, err := url.Parse("socks5://127.0.0.1:9050")
	errors.Check(err)

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	errors.Check(err)

	tbTransport := &http.Transport{Dial: tbDialer.Dial}
	return &http.Client{Transport: tbTransport}
}
