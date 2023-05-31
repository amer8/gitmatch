
# gitmatch

Matches a project folder to tags and/or commits of a git repository. It can be useful when you're migrating before Go 1.11 projects - which were managing their dependencies via $GOPATH - to Go Modules, and need to find out to which commit the content belongs.

**Stability**: Experimental

## Installation

```shell
go install github.com/amer8/gitmatch@latest
``` 

## Usage

```
gitmatch [--commits] [--tags] <repository-url>#<branch> <local-project-directory>
```

### Examples

```shell
gitmatch https://github.com/username/repo.git /var/folder

Found matching commit:  e5295cfa28272ada50cb9b5d97696ee06e49ed42  Date:  2015-11-21 11:10:40 +0800
```

If you want to match by tag or by commit only, use `--tags` or `--commits`.

```shell
gitmatch --tags https://github.com/username/repo.git /var/folder

Found matching tag:  1.1.0  Date:  2015-11-21 11:10:40 +0800
```

If you want to match commits on a specific branch provide `#branchname`.

```shell
gitmatch https://github.com/username/repo.git#branchname /var/folder
```
