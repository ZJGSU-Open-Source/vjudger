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

type PKUJudger struct {
	client   *http.Client
	token    string
	pat      *regexp.Regexp
	username string
	userpass string
}

const PKUToken = "PKU"

var PKURes = map[string]int{"Waiting": 0,
	"Compiling":             1,
	"Running & Judging":     1,
	"Compile Error":         2,
	"Accepted":              3,
	"Runtime Error":         4,
	"Wrong Answer":          5,
	"Time Limit Exceeded":   6,
	"Memory Limit Exceeded": 7,
	"Output Limit Exceeded": 8,
	"Presentation Error":    9,
	"System Error":          10}

var PKULang = map[int]int{
	LanguageNA:   -1,
	LanguageC:    1,
	LanguageCPP:  0,
	LanguageJAVA: 2}

func (h *PKUJudger) Init(_ UserInterface) error {
	jar, _ := cookiejar.New(nil)
	h.client = &http.Client{Jar: jar, Timeout: time.Second * 10}
	h.token = PKUToken
	pattern := `<tr align=center><td>(\d+)</td><td><a href=userstatus?user_id=vsake>vsake</a></td><td>.*?<font color=.*?>(.*?)</font>.*?</td><td>(.*?)</td><td>(.*?)</td><td><a href=showsource?solution_id=\d+ target=_blank>.*?</a></td><td>(\d+)B</td><td>(.*?)</td></tr>`
	//runid - result - memory - time - code_length - submit time
	h.pat = regexp.MustCompile(pattern)
	h.username = "vsake"
	h.userpass = "JC945312"
	return nil
}

func (h *PKUJudger) Match(token string) bool {
	if token == PKUToken {
		return true
	}
	return false
}

func (h *PKUJudger) Login(_ UserInterface) error {

	h.client.Get("http://poj.org/login")

	uv := url.Values{}
	uv.Add("user_id1", h.username)
	uv.Add("password1", h.userpass)
	uv.Add("B1", "login")
	uv.Add("url", ".")

	req, err := http.NewRequest("POST", "http://poj.org/login", strings.NewReader(uv.Encode()))
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
	// log.Println(html)
	if strings.Index(html, "Please retry after 100ms.Thank you.") >= 0 ||
		strings.Index(html, h.username) < 0 {
		return LoginFailed
	}

	return nil
}

func (h *PKUJudger) Submit(u UserInterface) error {

	uv := url.Values{}

	uv.Add("problem_id", strconv.Itoa(u.GetVid()))
	uv.Add("language", strconv.Itoa(PKULang[u.GetLang()]))
	uv.Add("source", u.GetCode())
	uv.Add("submit", "Submit")

	req, err := http.NewRequest("POST", "http://poj.org/submit", strings.NewReader(uv.Encode()))
	if err != nil {
		log.Println(err)
		return BadInternet
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	u.SetSubmitTime(time.Now())
	resp, err := h.client.Do(req)
	if err != nil {
		return BadInternet
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
	// log.Println(html)
	if strings.Index(html, "No such problem") >= 0 {
		return NoSuchProblem
	}
	if strings.Index(html, "Source code too long or too short,submit FAILED;") >= 0 {
		return SubmitFailed
	}

	return nil
}

func (h *PKUJudger) GetStatus(u UserInterface) error {

	statusUrl := "http://poj.org/status?problem_id=" +
		strconv.Itoa(u.GetVid()) + "&user_id=" +
		h.username + "&result=&language=" +
		strconv.Itoa(PKULang[u.GetLang()])
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
		log.Println(string(b))
		log.Println(AllStatus)
		return BadStatus
		layout := "2006-01-02 15:04:05 (MST)" //parse time
		for i := len(AllStatus) - 1; i >= 0; i-- {
			status := AllStatus[i]
			t, _ := time.Parse(layout, status[6]+" (CST)")
			log.Println(t, u.GetSubmitTime())
			// t = t.Add(40 * time.Second) //HDU server's time is less 36s.
			if t.After(u.GetSubmitTime()) {
				rid := status[1] //remote server run id
				u.SetResult(HDURes[status[2]])

				if u.GetResult() >= JudgeRJ {
					if u.GetResult() == JudgeCE {
						CE, err := h.GetCEInfo(rid)
						if err != nil {
							log.Println(err)
						}
						u.SetErrorInfo(CE)
					}
					Time, _ := strconv.Atoi(status[4][:len(status[4])-2])
					Mem, _ := strconv.Atoi(status[3][:len(status[3])-1])
					Length, _ := strconv.Atoi(status[5])
					u.SetResource(Time, Mem, Length)
					return nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (h *PKUJudger) GetCEInfo(rid string) (string, error) {
	resp, err := h.client.Get("http://poj.org/showcompileinfo?solution_id=" + rid)
	if err != nil {
		log.Println(err)
		return "", BadInternet
	}

	b, _ := ioutil.ReadAll(resp.Body)
	pre := `(?s)<font size="3">(.*?)</font>`
	re := regexp.MustCompile(pre)
	match := re.FindStringSubmatch(string(b))
	return match[1], nil
}

func (h *PKUJudger) Run(u UserInterface) error {
	for _, apply := range []func(UserInterface) error{h.Init, h.Login, h.Submit, h.GetStatus} {
		if err := apply(u); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
