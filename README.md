# go-commitlinter

go-commitlinter is simple commit message linter.

![Sample Image](.images/ss.png)

## Quick Start
```
go install github.com/masahiro331/go-commitlinter@0.0.1
echo "go-commitlinter" >> .git/hooks/commit-msg
chmod 755 .git/hooks/commit-msg
```

## Description

The go-commitlinter will detect and fail a commit message that is not in the following format.

```
<type>(<scope>): <subject>
```

The `type` and `scope` should always be lowercase as shown below.  
The `<scope>` can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted.

**Allowed `<type>` values:**
  - **feat** for a new feature for the user, not a new feature for build script.
  - **fix** for a bug fix for the user, not a fix to a build script.
  - **perf** for performance improvements.
  - **docs** for changes to the documentation.
  - **style** for formatting changes, missing semicolons, etc.
  - **refactor** for refactoring production code, e.g. renaming a variable.
  - **test** for adding missing tests, refactoring tests; no production code change.
  - **build** for updating build configuration, development tools or other changes irrelevant to the user.
  - **chore** for updates that do not apply to the above, such as dependency updates.

**`<scope>` example:**
  - parser
  - controller
  - some `package namespace`
  - etc...

## Other use cases
For example, if you want to validate the title of a pull request.  
Add the following github actions workflow.

```
name: Test
on:
  pull_request:
env:
  GO_VERSION: "1.17"
jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Go pull request message linter
        uses: masahiro331/go-commitlinter@0.1.0
        env:
          PR: ${{ github.event.number }}
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Custom Rules

This tool is able to custom rules.

```
go-commitlinter -rule rule.yaml
```

This is default rules.
```
skip_prefixes:
  - 'Merge branch '
  - 'BREAKING: '
type_rules:
  - type: feat
    description: for a new feature for the user, not a new feature for build script.
  - type: fix
    description: for a bug fix for the user, not a fix to a build script.
  - type: perf
    description: for performance improvements.
  - type: docs
    description: for changes to the documentation.
  - type: style
    description: for formatting changes, missing semicolons, etc.
  - type: refactor
    description: for refactoring production code, e.g. renaming a variable.
  - type: test
    description: for adding missing tests, refactoring tests; no production code change.
  - type: build
    description: for updating build configuration, development tools or other changes irrelevant to the user.
  - type: chore
    description: for updates that do not apply to the above, such as dependency updates.
reference: https://github.com/masahiro331/go-commitlinter#description
style_doc: The type and scope should always be lowercase.
scope_doc: The <scope> can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted.
```

- **skip_prefixes**: Use skip some titles. for example, merge commit "Merge branch 'main' of ....."
- **type_rules**: Use it to add your own type.
- **reference**: Include a link to the CONTRIBUTING GUILD.
- **style_doc**: Describe the specifications of style.
- **scope_doc**: Describe the specifications of scope.
