package main

import (
	"fmt"
	"os"
)

func panicOnError(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Abort: %s\n", err)
		fmt.Fprintf(os.Stderr, msg, args...)
		panic(err)
	}
}
