package util

import "fmt"

func Assert(exp bool, format string, vars ...interface{}) {
	if !exp {
		panic(fmt.Errorf(format, vars...))
	}
}