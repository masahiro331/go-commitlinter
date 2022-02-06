# go-commitlinter

go-commitlinter is simple commit message linter.

## Description

The go-commitlinter will detect and fail a commit message that is not in the following format.

```
<type>(<scope>): <subject>
```

The `type` and `scope` should always be lowercase as shown below.  
The `<scope>` can be empty (e.g. if the change is a global or difficult to assign to a single component), in which case the parentheses are omitted.


## How to use
```
go install github.com/masahiro331/go-commitlinter
echo "go-commitlinter" >> .git/hooks/commit-msg
```
