package http

import (
	"fmt"
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

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error %v ", resp.StatusCode)
	}

	return resp, err
}
