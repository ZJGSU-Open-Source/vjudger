package vjudger

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	// "os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ZJUJudger struct {
	client   *http.Client
	token    string
	pat      *regexp.Regexp
	username string
	userpass string
}

var ZJURes = map[string]int{"Queuing": 0,
	"Compile Error":         2,
	"Accepted":              3,
	"Segmentation Fault":    4,
	"Floating Point Error":  4,
	"Runtime Error":         4,
	"Wrong Answer":          5,
	"Time Limit Exceeded":   6,
	"Memory Limit Exceeded": 7,
	"Output Limit Exceeded": 8,
	"Presentation Error":    9}

var ZJULang = map[int]int{
	LanguageNA:   0,
	LanguageC:    1,
	LanguageCPP:  2,
	LanguageJAVA: 4}

func (h *ZJUJudger) Init(_ *User) error {
	jar, _ := cookiejar.New(nil)
	h.client = &http.Client{Jar: jar}
	h.token = "ZJU"
	//To fix
	pattern := `(\d+)</td><td>(.*?)</td><td>(?s:.*?)<font color=.*?>(.*?)</font>.*?<td>(\d+)MS</td><td>(\d+)K</td><td><a href="/viewcode.php\?rid=\d+"  target=_blank>(\d+) B</td><td>(.*?)</td>`
	h.pat = regexp.MustCompile(pattern)
	h.username = "mysake"
	h.userpass = "JC945312"
	return nil
}

func (h *ZJUJudger) Match(token string) bool {
	if token == h.token {
		return true
	}
	return false
}
func (h *ZJUJudger) Login(_ *User) error {

	h.client.Get("http://acm.zju.edu.cn/onlinejudge/login.do")

	uv := url.Values{}
	uv.Add("handle", h.username)
	uv.Add("password", h.userpass)

	req, err := http.NewRequest("POST", "http://acm.zju.edu.cn/onlinejudge/login.do", strings.NewReader(uv.Encode()))
	if err != nil {
		return BadInternet
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.Do(req)
	if err != nil {
		log.Println("err", err)
		return BadInternet
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)

	if strings.Index(html, "Handle or password is invalid.") >= 0 ||
		strings.Index(html, "Handle is required.") >= 0 ||
		strings.Index(html, "Password is required.") >= 0 {
		return LoginFailed
	}

	return nil
}

func (h *ZJUJudger) Submit(u *User) error {

	uv := url.Values{}
	uv.Add("problemId", strconv.Itoa(u.Vid))
	uv.Add("languageId", strconv.Itoa(ZJULang[u.Lang]))
	uv.Add("source", u.Code)

	req, err := http.NewRequest("POST", "http://acm.zju.edu.cn/onlinejudge/submit.do", strings.NewReader(uv.Encode()))
	if err != nil {
		log.Println(err)
		return BadInternet
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	u.SubmitTime = time.Now()
	resp, err := h.client.Do(req)
	if err != nil {
		return BadInternet
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
	if strings.Index(html, "No such problem.") >= 0 {
		return NoSuchProblem
	}

	return nil
}

func (h *ZJUJudger) GetStatus(u *User) error {

	statusUrl := "http://acm.zju.edu.cn/onlinejudge/showRuns.do?contestId=1" +
		"&problemCode=" + strconv.Itoa(u.Vid) +
		"&handle=" + h.username +
		"&languageIds=" + strconv.Itoa(u.Lang)

	endTime := time.Now().Add(MAX_WaitTime * time.Second)

	for true {
		if time.Now().After(endTime) {
			return BadStatus
		}
		resp, err := h.client.Get(statusUrl)
		if err != nil {
			return BadInternet
		}
		defer resp.Body.Close()

		b, _ := ioutil.ReadAll(resp.Body)
		AllStatus := h.pat.FindAllStringSubmatch(string(b), -1)

		layout := "2006-01-02 15:04:05 (MST)" //parse time
		for i := len(AllStatus) - 1; i >= 0; i-- {
			status := AllStatus[i]
			t, _ := time.Parse(layout, status[2]+" (CST)")
			// t = t.Add(40 * time.Second) //HDU server's time is less 36s.
			if t.After(u.SubmitTime) {
				rid := status[1] //remote server run id
				u.Result = HDURes[status[3]]

				if u.Result >= JudgeRJ {
					if u.Result == JudgeCE {
						u.CE, err = h.GetCEInfo(rid)
						if err != nil {
							log.Println(err)
						}
					}

					u.Time, _ = strconv.Atoi(status[4])
					u.Mem, _ = strconv.Atoi(status[5])
					u.Length, _ = strconv.Atoi(status[6])
					return nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (h *ZJUJudger) GetCEInfo(rid string) (string, error) {
	resp, err := h.client.Get("http://acm.zju.edu.cn/onlinejudge/showJudgeComment.do?submissionId=" + rid)
	if err != nil {
		log.Println(err)
		return "", BadInternet
	}

	b, _ := ioutil.ReadAll(resp.Body)
	return string(b), nil
}

func (h *ZJUJudger) Run(u *User) error {
	for _, apply := range []func(*User) error{h.Init, h.Login, h.Submit, h.GetStatus} {
		if err := apply(u); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
