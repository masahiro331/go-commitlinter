package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

const (
	commitMsgFilePath = ".git/COMMIT_EDITMSG"
	formatDoc         = "<type>(<scope>): <subject>"
	scopeDoc          = "The <scope> can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted."
	styleDoc          = "The <type> and <scope> should always be lowercase."
	subjectDoc        = "The first letter of <subject> should be lowercase."
)

func textRed(s string) string {
	return fmt.Sprintf("\033[0;31m%s\033[0m", s)
}

func textBrightGreen(s string) string {
	return fmt.Sprintf("\033[0;92m%s\033[0m", s)
}

func textBrightYellow(s string) string {
	return fmt.Sprintf("\033[0;93m%s\033[0m", s)
}

var (
	r = flag.String("rule", "", "select rule file path (config.yaml)")

	FormatRegularPattern = `([a-zA-Z]+)(\(.*\))?:\s+(.*)`

	errorTitle    = "============================ Invalid Message ================================"
	errorTemplate = "\n%s\ntitle message:	%s\ncorrect format:	%s\n\n%s\n\nSee: %s\n"
	footer        = "============================================================================="

	ErrStyle   = errors.New("invalid style error")
	ErrType    = errors.New("invalid type error")
	ErrFormat  = errors.New("invalid format error")
	ErrScope   = errors.New("invalid scope error")
	ErrSubject = errors.New("invalid subject error")

	DefaultConfig = Config{
		SkipPrefixes: []string{
			"Merge branch ",
			"BREAKING: ",
		},
		Reference: "https://github.com/masahiro331/go-commitlinter#description",
		TypeRules: TypeRules{
			{
				Type:        "feat",
				Description: "for a new feature for the user, not a new feature for build script.",
			},
			{
				Type:        "fix",
				Description: "for a bug fix for the user, not a fix to a build script.",
			},
			{
				Type:        "perf",
				Description: "for performance improvements.",
			},
			{
				Type:        "docs",
				Description: "for changes to the documentation.",
			},
			{
				Type:        "style",
				Description: "for formatting changes, missing semicolons, etc.",
			},
			{
				Type:        "refactor",
				Description: "for refactoring production code, e.g. renaming a variable.",
			},
			{
				Type:        "test",
				Description: "for adding missing tests, refactoring tests; no production code change.",
			},
			{
				Type:        "build",
				Description: "for updating build configuration, development tools or other changes irrelevant to the user.",
			},
			{
				Type:        "chore",
				Description: "for updates that do not apply to the above, such as dependency updates.",
			},
		},
		StyleDoc: styleDoc,
		ScopeDoc: scopeDoc,
	}
)

type TypeRule struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

type TypeRules []TypeRule

func (typeRules TypeRules) String() string {
	ret := "Allowed <type> values\n"
	for _, tr := range typeRules {
		ret += fmt.Sprintf("%s\t%s\n", textBrightYellow(tr.Type), tr.Description)
	}

	return ret
}

type Config struct {
	SkipPrefixes []string  `yaml:"skip_prefixes"`
	TypeRules    TypeRules `yaml:"type_rules"`
	Reference    string    `yaml:"reference"`
	StyleDoc     string    `yaml:"style_doc"`
	ScopeDoc     string    `yaml:"scope_doc"`
}

type Format struct {
	Type    string
	Scope   string
	Subject string
}

func NewConfig(filepath string) (Config, error) {
	if filepath == "" {
		return DefaultConfig, nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return Config{}, xerrors.Errorf("failed to open config: %w", err)
	}
	var conf Config
	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		return Config{}, xerrors.Errorf("failed to parse yaml: %w", err)
	}

	return conf, nil
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

func (f Format) subjectLinter() error {
	if !(len(f.Subject) > 0) {
		return ErrFormat
	}
	r := rune(f.Subject[0])
	if unicode.IsUpper(r) {
		return ErrSubject
	}

	return nil
}

func (f Format) typeLinter(c Config) error {
	for _, r := range c.TypeRules {
		if r.Type == f.Type {
			return nil
		}
	}
	if f.Type != strings.ToLower(f.Type) {
		return ErrStyle
	}

	return ErrType
}

func (f Format) Verify(c Config) error {
	if err := f.typeLinter(c); err != nil {
		return err
	}

	if err := f.scopeLinter(); err != nil {
		return err
	}

	if err := f.subjectLinter(); err != nil {
		return err
	}

	return nil
}

func run() (string, Config, error) {
	flag.Parse()

	conf, err := NewConfig(*r)
	if err != nil {
		log.Fatal(err)
	}

	s, err := getMessage()
	if err != nil {
		return "", conf, err
	}
	for _, skipPrefix := range conf.SkipPrefixes {
		if strings.HasPrefix(s, skipPrefix) {
			return "", conf, nil
		}
	}

	format, err := NewFormat(s)
	if err != nil {
		return s, conf, err
	}
	if err := format.Verify(conf); err != nil {
		return s, conf, err
	}

	return "", conf, nil
}

func finally(m string, conf Config, err error) {
	message := ""
	switch err {
	case ErrFormat, ErrType:
		message = fmt.Sprintf(errorTemplate, textRed(errorTitle), textRed(m), textBrightGreen(formatDoc), conf.TypeRules, textBrightGreen(conf.Reference))
	case ErrStyle:
		message = fmt.Sprintf(errorTemplate, textRed(errorTitle), textRed(m), textBrightGreen(formatDoc), textBrightYellow(conf.StyleDoc), textBrightGreen(conf.Reference))
	case ErrScope:
		message = fmt.Sprintf(errorTemplate, textRed(errorTitle), textRed(m), textBrightGreen(formatDoc), textBrightYellow(conf.ScopeDoc), textBrightGreen(conf.Reference))
	case nil:
		return
	default:
		log.Fatal(xerrors.Errorf("unspecified error: %w", err))
	}
	message = fmt.Sprintf("%s\n%s", message, textRed(footer))
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}

func getMessage() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	b, _, _ := reader.ReadLine()
	if len(b) != 0 {
		return string(b), nil
	}

	f, err := os.Open(commitMsgFilePath)
	if err != nil {
		return "", err
	}

	reader = bufio.NewReader(f)
	b, _, err = reader.ReadLine()
	if err != nil {
		return "", ErrFormat
	}

	return string(b), nil
}

func main() {
	finally(run())
}
