package vjudger

import (
	"time"
)

type User struct {
	Uid string //uer identify
	Sid int    //solution id
	Vid int    //remote problem id

	OJ     string //OJ Token
	Result int    //Judge result
	CE     string // Compile error information
	Code   string //uesr code
	Time   int    //user time
	Mem    int    //user memory
	Lang   int    //user languaga
	Length int    //user code length

	ErrorCode  int       //remote vjudger error
	SubmitTime time.Time //time that client submit
}

const MAX_WaitTime = 120

type Vjudger interface {
	Init(*User) error
	Login(*User) error
	Submit(*User) error
	GetStatus(*User) error
	Run(*User) error
	Match(string) bool
}

func (u *User) NewSolution() {

}

func (u *User) UpdateSolution() {

}

var VJs = []Vjudger{&HDUJudger{}}

func Judge(u *User) {
	for _, vj := range VJs {
		if vj.Match(u.OJ) {
			vj.Run(u)
			break
		}
	}
}
