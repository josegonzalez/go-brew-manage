# go-brew-manage [![CircleCI](https://circleci.com/gh/josegonzalez/go-brew-manage.svg?style=svg)](https://circleci.com/gh/josegonzalez/go-brew-manage)

A tool for managing a homebrew installation.

## Installation

Install it using the "go get" command:

    go get github.com/josegonzalez/go-brew-manage

## Usage

This binary depends on a `brew.yaml` file, and is invoked like so:

```shell
# if no config path is specified, a `brew.yaml` in the current directory is used
brew-manage 

# a `brew.yaml` path can be specified via the `-config` flag
brew-manage -config brew.yaml
```

### brew.yaml format

The `brew.yaml` format is a list of "calls" for homebrew to execute. Each call is of the following form:

```yaml
- CALL_TYPE:
  name: NAME_OF_ITEM_MANAGED
  state: <present|absent>
```

> At this time, the `state` parameter has no meaning and is assumed to be `present`.

#### homebrew_tap

Homebrew taps can be managed by using the `homebrew_tap` type.

```yaml
- homebrew_tap:
  name: hakamadare/goenv
```

#### homebrew_cask

homebrew casks can be managed by using the `homebrew_cask` type.

```yaml
- homebrew_cask:
  name: docker
```

Using the `homebrew_cask` type will automatically add the following casks to your installation list:

- homebrew/cask
- homebrew/cask-drivers
- homebrew/cask-fonts
- homebrew/cask-versions

Cask installation may occasionally require root permissions. In these cases, you may see a password prompt during `brew-manage` invocation.

#### homebrew_formula

Homebrew formula can be managed by using the `homebrew_formula` type.

```yaml
- homebrew_formula:
  name: go
```

#### homebrew_pip

Python pip packages can be installed via homebrew by using the `homebrew_pip` type.

```yaml
- homebrew_pip:
  name: dotfiles
```

Using the `homebrew_pip` type will automatically install the `brew-pip` and `python` formulae, as it depends on those in order for packages to be installed correctly. Please note that these packages will therefore be installed via `python3`, as that is the version in use for the official `python` formula at this time.