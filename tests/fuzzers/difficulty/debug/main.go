package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fgeth/fgeth/tests/fuzzers/difficulty"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: debug <file>")
		os.Exit(1)
	}
	crasher := os.Args[1]
	data, err := ioutil.ReadFile(crasher)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading crasher %v: %v", crasher, err)
		os.Exit(1)
	}
	difficulty.Fuzz(data)
}
