package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

const (
	commitMsgFilePath = ".git/COMMIT_EDITMSG"
	formatDoc         = "<type>(<scope>): <subject>"
	errorTitle        = "\033[0;31m============================ Invalid Commit Message ================================\033[0m"
	errorTemplate     = "\n%s\ncommit message:	\033[0;31m%s\033[0m\ncorrect format:	\033[0;92m%s\033[0m\n\n%s\n%s\n\n"
	footer            = "\033[0;31m====================================================================================\033[0m"
)

var (
	FormatRegularPattern = `([a-zA-Z]+)(\(.*\))?:\s+(.*)`
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
	scopeDoc = "\033[0;93mThe <scope> can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted.\033[0m"
	styleDoc = "\033[0;93mThe type and scope should always be lowercase.\033[0m"

	ErrStyle  = errors.New("invalid style error")
	ErrType   = errors.New("invalid type error")
	ErrFormat = errors.New("invalid format error")
	ErrScope  = errors.New("invalid scope error")
)

type Format struct {
	Type    string
	Scope   string
	Subject string
}

func NewFormat(m string) (Format, error) {
	p, err := regexp.Compile(FormatRegularPattern)
	if err != nil {
		return Format{}, err
	}
	ss := p.FindAllStringSubmatch(m, 1)
	if len(ss) == 0 || len(ss[0]) != 4 {
		return Format{}, ErrFormat
	}

	t := ss[0][1]
	subject := ss[0][3]
	if t == "" || subject == "" {
		return Format{}, ErrFormat
	}

	scope := ss[0][2]
	if scope != "" {
		scope = strings.TrimPrefix(strings.TrimSuffix(scope, ")"), "(")
		if scope == "" {
			return Format{}, ErrScope
		}
	}

	f := Format{
		Type:    t,
		Scope:   scope,
		Subject: subject,
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

func (f Format) Verify() error {
	if err := f.typeLinter(); err != nil {
		return err
	}

	if err := f.scopeLinter(); err != nil {
		return err
	}
	return nil
}

func run() (string, error) {
	f, err := os.Open(commitMsgFilePath)
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(f)
	b, _, err := r.ReadLine()
	if err != nil {
		return "", ErrFormat
	}

	format, err := NewFormat(string(b))
	if err != nil {
		return string(b), err
	}
	if err := format.Verify(); err != nil {
		return string(b), err
	}

	return "", nil
}

func finally(m string, err error) {
	message := ""
	switch err {
	case ErrFormat, ErrType:
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, typeDoc, footer)
	case ErrStyle:
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, styleDoc, footer)
	case ErrScope:
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, scopeDoc, footer)
	case nil:
		return
	default:
		log.Fatal(xerrors.Errorf("unspecified error: %w", err))
	}
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}

func main() {
	finally(run())
}
