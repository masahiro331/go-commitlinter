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

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

const (
	commitMsgFilePath = ".git/COMMIT_EDITMSG"
	formatDoc         = "<type>(<scope>): <subject>"
	errorTitle        = "\033[0;31m============================ Invalid Commit Message ================================\033[0m"
	errorTemplate     = "\n%s\ncommit message:	\033[0;31m%s\033[0m\ncorrect format:	\033[0;92m%s\033[0m\n\n%s\n%s\n\n"
	footer            = "\033[0;31m====================================================================================\033[0m"
)

var (
	r = flag.String("rule", "", "select rule file path (config.yaml)")

	FormatRegularPattern = `([a-zA-Z]+)(\(.*\))?:\s+(.*)`

	scopeDoc = "\033[0;93mThe <scope> can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted.\033[0m"
	styleDoc = "\033[0;93mThe type and scope should always be lowercase.\033[0m"

	ErrStyle  = errors.New("invalid style error")
	ErrType   = errors.New("invalid type error")
	ErrFormat = errors.New("invalid format error")
	ErrScope  = errors.New("invalid scope error")

	DefaultRules = Config{
		SkipPrefixes: []string{
			"Merge branch ",
		},
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
	}
)

type TypeRule struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

type TypeRules []TypeRule

func (typeRules TypeRules) String() string {
	ret := "Allows type values\n"
	for _, tr := range typeRules {
		ret += fmt.Sprintf("\033[0;93m%s\033[0m\t%s\n", tr.Type, tr.Description)
	}

	return ret
}

type Config struct {
	SkipPrefixes []string  `yaml:"skip_prefixes"`
	TypeRules    TypeRules `yaml:"type_rules"`
}

type Format struct {
	Type    string
	Scope   string
	Subject string
}

type Linter struct {
	Conf   Config
	Format Format
}

func NewConfig(filepath string) (Config, error) {
	if filepath == "" {
		return DefaultRules, nil
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

func finally(m string, conf Config, err error) {
	message := ""
	switch err {
	case ErrFormat, ErrType:
		message = fmt.Sprintf(errorTemplate, errorTitle, m, formatDoc, conf.TypeRules, footer)
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
