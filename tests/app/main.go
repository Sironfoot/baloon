package main

import (
	"flag"
	"fmt"
)

func main() {
	var readyMessage string
	flag.StringVar(&readyMessage, "ready_statement", "", "")
	flag.Parse()

	fmt.Println(readyMessage)
}
