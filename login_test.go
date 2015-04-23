package vjudger

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"
)

func Test_Login(t *testing.T) {
	h := &HDUJudger{}
	jar, _ := cookiejar.New(nil)
	h.client = &http.Client{Jar: jar, Timeout: time.Second * 10}

	h.username = "mysake"
	h.userpass = "123456"

	h.client.Get("http://acm.hdu.edu.cn")

	uv := url.Values{}
	uv.Add("username", h.username)
	uv.Add("userpass", h.userpass)
	uv.Add("login", "Sign In")

	req, err := http.NewRequest("POST", "http://acm.hdu.edu.cn/userloginex.php?action=login", strings.NewReader(uv.Encode()))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8,zh-CN;q=0.6,zh;q=0.4,zh-TW;q=0.2")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Length", "45")
	req.Header.Add("Cookie", "exesubmitlang=0; PHPSESSID=lr3h2jm5k64gvr240lfnvp03t0; CNZZDATA1254072405=1250590032-1421111964-%7C1425188228")
	req.Header.Add("DNT", "1")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Origin", "http://acm.hdu.edu.cn")
	req.Header.Add("Referer", "http://acm.hdu.edu.cn/userloginex.php?action=login")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.115 Safari/537.36")
	resp, err := h.client.Do(req)
	if err != nil {
		log.Println("err", err)
		return
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
	if strings.Index(html, "No such user or wrong password.") >= 0 {
		log.Println("No such")
	}

}
