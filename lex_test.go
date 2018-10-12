package robots

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestLex(t *testing.T) {
	f, err := os.Open("testdata/pathological.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	for _, v := range lex(string(buf)) {
		if v.typ == 0 {
			fmt.Println(v.typ)
			continue
		}
		fmt.Println(v.typ, v.val)
	}
}
