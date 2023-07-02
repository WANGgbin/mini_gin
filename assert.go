package mini_gin

import "fmt"

func assert(exp bool, format string, vars ...interface{}) {
	if !exp {
		panic(fmt.Errorf(format, vars...))
	}
}