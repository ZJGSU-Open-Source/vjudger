package vjudger

import (
	"time"
)

type UserInterface interface {
	GetResult() int
	SetResult(int) int
	SetResource(int, int, int)
	SetErrorInfo(string)
	GetSubmitTime() time.Time
	SetSubmitTime(time.Time)
	GetCode() string
	GetOJ() string
	GetLang() int
	GetVid() int
	UpdateSolution()
}

const MAX_WaitTime = 120

type Vjudger interface {
	Init(UserInterface) error
	Login(UserInterface) error
	Submit(UserInterface) error
	GetStatus(UserInterface) error
	Run(UserInterface) error
	Match(string) bool
}

var VJs = []Vjudger{&HDUJudger{}}

func Judge(u UserInterface) {
	for _, vj := range VJs {
		if vj.Match(u.GetOJ()) { //init?match
			vj.Run(u)
			break
		}
	}
}
