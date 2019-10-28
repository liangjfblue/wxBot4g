package wcbot

import (
	"net/http"
	"testing"
)

func TestTextHandle(t *testing.T) {
	urlStr := "http://127.0.0.1:7788/v1/msg/text?to=测试&word=那就等下&appKey=khr9348yo1oh"
	_, err := http.Get(urlStr)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("send text msg ok")
}
