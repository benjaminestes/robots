package robots

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("testdata/pathological.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	items := lex(string(buf))
	fmt.Println(parse(items))
}
