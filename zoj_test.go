package vjudger

import (
	"log"
	"testing"
)

func Test_ZJU(t *testing.T) {
	u := &User{Vid: 1, Lang: 2}
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
	// z := &ZJUJudger{}
	// z.Init(u)
	// z.Login(u)
	// z.Submit(u)
	log.Println(*u)
}
