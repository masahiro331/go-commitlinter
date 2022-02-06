package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	commitMsgFilePath = ".git/COMMIT_EDITMSG"
	styleDoc          = `The type and scope should always be lowercase.`
	formatDoc         = "<type>(<scope>): <subject>"
	errorTitle        = "\033[0;31m============================ Invalid Commit Message ================================\033[0m"
	errorTemplate     = "\n%s\ncommit message:	\033[0;31m%s\033[0mcorrect format:	\033[0;92m%s\033[0m\n\n%s\n%s\n\n"
	footer            = "\033[0;31m====================================================================================\033[0m"
)

var (
	FormatRegularPattern = `([a-zA-Z]+)\((.+)\):\s(.*)`
	requiredTypes        = []string{
		"feat",
		"fix",
		"perf",
		"docs",
		"style",
		"refactor",
		"test",
		"build",
		"chore",
	}
	typeDoc = "Allows type values\n" +
		"\033[0;93mfeat\033[0m:		for a new feature for the user, not a new feature for build script.\n" +
		"\033[0;93mfix\033[0m:		for a bug fix for the user, not a fix to a build script. \n" +
		"\033[0;93mperf\033[0m:		for performance improvements.\n" +
		"\033[0;93mdocs\033[0m:		for changes to the documentation.\n" +
		"\033[0;93mstyle\033[0m:		for formatting changes, missing semicolons, etc.\n" +
		"\033[0;93mrefactor\033[0m:	for refactoring production code, e.g. renaming a variable.\n" +
		"\033[0;93mtest\033[0m:		for adding missing tests, refactoring tests; no production code change.\n" +
		"\033[0;93mbuild\033[0m:		for updating build configuration, development tools or other changes irrelevant to the user.\n" +
		"\033[0;93mchore\033[0m:		for updates that do not apply to the above, such as dependency updates."

	ErrStyle  = errors.New("invalid style error")
	ErrType   = errors.New("invalid type error")
	ErrFormat = errors.New("invalid format error")
)

type Format struct {
	Type    string
	Scope   string
	Subject string
}

func messageParser(m string) (Format, error) {
	p, err := regexp.Compile(FormatRegularPattern)
	if err != nil {
		log.Fatal(err)
	}
	ss := p.FindAllStringSubmatch(m, 1)
	if len(ss) == 0 || len(ss[0]) != 4 {
		return Format{}, ErrFormat
	}

	f := Format{
		Type:    ss[0][1],
		Scope:   ss[0][2],
		Subject: ss[0][3],
	}
	if f.Type == "" || f.Scope == "" || f.Subject == "" {
		return Format{}, ErrFormat
	}
	return f, nil
}

func (f Format) scopeLinter() error {
	if f.Scope != strings.ToLower(f.Scope) {
		return ErrStyle
	}

	return nil
}
func (f Format) typeLinter() error {
	for _, t := range requiredTypes {
		if t == f.Type {
			return nil
		}
	}
	if f.Type != strings.ToLower(f.Type) {
		return ErrStyle
	}

	return ErrType
}

func run() (string, error) {
	b, err := os.ReadFile(commitMsgFilePath)
	if err != nil {
		log.Fatal(err)
	}
	s := string(b)
	format, err := messageParser(s)
	if err != nil {
		return s, err
	}

	if err := format.typeLinter(); err != nil {
		return s, err
	}

	if err := format.scopeLinter(); err != nil {
		return s, err
	}

	return "", nil
}

func finally(m string, err error) {
	message := ""
	if err != nil {
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, typeDoc, footer)
	}
	if errors.Is(ErrFormat, err) {
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, typeDoc, footer)
	}
	if errors.Is(ErrStyle, err) {
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, styleDoc, footer)
	}
	if errors.Is(ErrType, err) {
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, typeDoc, footer)
	}
	fmt.Println(message)
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	finally(run())
}
