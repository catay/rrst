package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	tmpSuffix = ".filepart"
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

// Download file from URL and save to specified path.
func HttpGetFile(url, filepath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := HttpProxyGet(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	f, err := os.Create(filepath + tmpSuffix)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	err = os.Rename(filepath+tmpSuffix, filepath)
	if err != nil {
		return err
	}

	return nil
}
