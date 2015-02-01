package vjudger

import (
	"log"
	"testing"
	// "time"
)

func Test_PKU(t *testing.T) {
	u := &User{Vid: 1000, Lang: LanguageCPP}
	u.Code = `
#include<iostream>
 
using namespace std;
 
int main(){
   int a,b;
   while(cin>>a>>b){
      cout<<a+b<<endl;
   }
   return 0;
}
	`
	h := &PKUJudger{}
	err := h.Run(u)
	if err != nil {
		t.Error(err)
	}
	log.Println(*u)
}
