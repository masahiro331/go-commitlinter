package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

const (
	commitMsgFilePath = ".git/COMMIT_EDITMSG"
)

var (
	FormatRegularPattern = `([a-zA-Z]+)\((.+)\):\s(.*)`
)

type Format struct {
	Type    string
	Scope   string
	Subject string
}

func messageParser(m string) {

}

func main() {
	fmt.Println("pre-commit !!!!  binary message")
	p, err := regexp.Compile(FormatRegularPattern)
	if err != nil {
		log.Fatal(err)
	}
	b, err := os.ReadFile(commitMsgFilePath)
	if err != nil {
		log.Fatal(err)
	}

	bbb := p.FindAllSubmatch(b, 1)
	fmt.Println(bbb)

	os.Exit(1)
}
