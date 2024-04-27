package bd

import (
	"runtime"
	"strings"
	"fmt"
)

func Assert(cond bool) {
	if !cond {
		fname := getPos()
		fmt.Printf("Assertion failed at %s\n", fname)
		panic("assertion failed")
	}
}

func getPos() string {
	fname := ""
	pc, _, line, _ := runtime.Caller(2)
	fc := runtime.FuncForPC(pc)
	if fc != nil {
		tmp := fc.Name()
		words := strings.Split(tmp, ".")
		fname = fmt.Sprintf("%s:%d", words[len(words)-1], line)
	} else {
		fname = fmt.Sprintf("Null(): %d", line)
	}

	return fname
}

func pl() {
	fmt.Printf("run at %s\n", getPos())
}


