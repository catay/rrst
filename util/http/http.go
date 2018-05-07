package http

import (
	"net/http"
)

// Wrapper function taking a proxy into account when doing an HTTP request.
func HttpProxyGet(req *http.Request) (resp *http.Response, err error) {
	proxyUrl, err := http.ProxyFromEnvironment(req)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{Transport: tr}

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}
