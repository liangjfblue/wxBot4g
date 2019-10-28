package httpClient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type Client struct {
	cookies []*http.Cookie
	headers map[string]string
	client  http.Client
}

func New(headers map[string]string) *Client {
	httpClient := new(Client)
	httpClient.cookies = make([]*http.Cookie, 0)
	httpClient.headers = make(map[string]string)
	for k, v := range headers {
		httpClient.headers[k] = v
	}

	return httpClient
}

func (h *Client) Post(url string, body interface{}) ([]byte, error) {
	var (
		err error
		req *http.Request
	)
	if body != nil {
		jBody, err := json.Marshal(body)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		req, err = http.NewRequest("POST", url, strings.NewReader(string(jBody)))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}

	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if len(h.cookies) > 0 {
		for _, cookie := range h.cookies {
			req.AddCookie(cookie)
		}
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	return out, nil
}

func (h *Client) PostString(url string, body string) ([]byte, error) {
	var (
		err error
		req *http.Request
	)

	req, err = http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if len(h.cookies) > 0 {
		for _, cookie := range h.cookies {
			req.AddCookie(cookie)
		}
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	return out, nil
}

func (h *Client) PostMedia(url string, body []byte) ([]byte, error) {
	var (
		err error
		req *http.Request
	)
	if body != nil {
		req, err = http.NewRequest("POST", url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}

	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if len(h.cookies) > 0 {
		for _, cookie := range h.cookies {
			req.AddCookie(cookie)
		}
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	return out, nil
}

func (h *Client) Get(url string, params url.Values) ([]byte, error) {
	if params != nil && len(params) > 0 {
		url = url + params.Encode()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if len(h.cookies) > 0 {
		for _, cookie := range h.cookies {
			req.AddCookie(cookie)
		}
	}
	if h.headers != nil {
		for k, v := range h.headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if len(h.cookies) == 0 {
		h.SetCookie(resp.Cookies())
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	return out, nil
}

func (h *Client) SetCookie(cookie []*http.Cookie) {
	h.cookies = cookie
}

func (h *Client) GetCookie() []*http.Cookie {
	return h.cookies
}

func (h *Client) GetCookieByName(name string) *http.Cookie {
	for _, cookie := range h.cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func (h *Client) SetHeader(header map[string]string) {
	if header != nil {
		for key, value := range header {
			h.headers[key] = value
		}
	}
}

func (h *Client) GetHeader() map[string]string {
	return h.headers
}

func (h *Client) DelHeader(header map[string]string) {
	if header != nil {
		for key, _ := range header {
			delete(h.headers, key)
		}
	}
}
