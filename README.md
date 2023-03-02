# lazynpm

![CI](https://github.com/jesseduffield/lazynpm/workflows/Continuous%20Integration/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jesseduffield/lazynpm)](https://goreportcard.com/report/github.com/jesseduffield/lazynpm) [![GolangCI](https://golangci.com/badges/github.com/jesseduffield/lazynpm.svg)](https://golangci.com) [![GoDoc](https://godoc.org/github.com/jesseduffield/lazynpm?status.svg)](http://godoc.org/github.com/jesseduffield/lazynpm) [![GitHub tag](https://img.shields.io/github/tag/jesseduffield/lazynpm.svg)]() [![TODOs](https://badgen.net/https/api.tickgit.com/badgen/github.com/jesseduffield/lazynpm)](https://www.tickgit.com/browse?repo=github.com/jesseduffield/lazynpm)

A simple terminal UI for npm commands, written in Go with the [gocui](https://github.com/jroimartin/gocui "gocui") library.

npm is pretty cool, but some of its workflows are a little too much for somebody with my atrocious short term memory. If I need to link a couple of dependencies to a package I need to do an `npm install` on the package so we're up to date, then cd to each dependency package and `npm install`, then `npm run build`, then `npm link` if it's not already globally linked (and how would I know that it is?) and then cd back to the original package and do `npm link <dependency>` for each dependency. Pretty much every time I'll get halfway through, suspect that I've forgotten a step, then start again from scratch. But who needs a functionining brain when have a tool where every step in the process take one keypress and at a glance you can see how everything is linked up?

lazynpm is the younger brother of lazygit and lazydocker, and has learnt from both its siblings how best to make life lazier in the terminal so that you can focus on what matters: programming.

![Gif](/docs/resources/demo2.gif)

## Table of contents

- [Installation](#installation)
  - [Binary releases](#binary-releases)
  - [Homebrew](#homebrew)
  - [Go](#go)
- [Usage](#usage)
  - [Keybindings](#keybindings)
  - [Changing directory on exit](#changing-directory-on-exit)
- [Configuration](#configuration)
- [Tutorials](#tutorials)
- [Cool Features](#cool-features)
- [Contributing](#contributing)
- [Donate](#donate)

Github Sponsors is matching all donations dollar-for-dollar for 12 months so if you're feeling generous consider [sponsoring me](https://github.com/sponsors/jesseduffield)

[<img src="https://i.imgur.com/CXSiCu1.jpg">](https://www.youtube.com/watch?v=J-FJdxrESqw)

## Installation

This program is not compatible with Windows because one of its dependencies, pty, is not compatible.

### Binary Releases

For Mac OS or Linux, you can download a binary release [here](../../releases).

### Homebrew

Normally the lazynpm formula can be found in the Homebrew core but we suggest you tap our formula to get the frequently updated one. It works with Linux, too.

Tap:

```
brew install jesseduffield/lazynpm/lazynpm
```

### Go

```sh
go install github.com/jesseduffield/lazynpm@latest
```

Please note:
If you get an error claiming that lazynpm cannot be found or is not defined, you
may need to add `~/go/bin` to your \$PATH (MacOS/Linux). Not to be mistaken for `$GOROOT/bin` (which is for Go's own binaries,
not apps like lazynpm).


## Usage

Call `lazynpm` in your terminal inside a git repository.

```sh
$ lazynpm
```

If you want, you can
also add an alias for this with `echo "alias lzn='lazynpm'" >> ~/.zshrc` (or
whichever rc file you're using).

### Keybindings

You can check out the list of keybindings [here](/docs/keybindings).

### Changing Directory On Exit

If you change repos in lazynpm and want your shell to change directory into that repo on exiting lazynpm, add this to your `~/.zshrc` (or other rc file):

```
lzn()
{
    export LAZYNPM_NEW_DIR_FILE=~/.lazynpm/newdir

    lazynpm "$@"

    if [ -f $LAZYNPM_NEW_DIR_FILE ]; then
            cd "$(cat $LAZYNPM_NEW_DIR_FILE)"
            rm -f $LAZYNPM_NEW_DIR_FILE > /dev/null
    fi
}
```

Then `source ~/.zshrc` and from now on when you call `lzn` and exit you'll switch directories to whatever you were in inside lazyigt. To override this behaviour you can exit using `shift+Q` rather than just `q`.

## Configuration

Check out the [configuration docs](docs/Config.md).

## Tutorials

- [Video Tutorial](https://www.youtube.com/watch?v=J-FJdxrESqw)
- [Twitch Stream](https://www.twitch.tv/jesseduffield)

## Cool features

- easily link packages and see which packages are linked
- pack packages and install from tarballs
- instantly know which dependencies are behind (or ahead) based on semver
- install/update multiple things at a time
- view at a glance each description of a package's dependencies
- easily switch between packages
- easily change the version constraints on packages
- easily add/remove/modify dependencies and scripts

## Contributing

I've written the code so that it's easy to build upon, so contributors are welcome! Please check out the [contributing guide](CONTRIBUTING.md).
For contributor discussion about things not better discussed here in the repo, join the [discord](https://discord.gg/ehwFt2t4wt) channel

## Donate

If you would like to support the development of lazynpm, consider [sponsoring me](https://github.com/sponsors/jesseduffield) (github is matching all donations dollar-for-dollar for 12 months)

## Work in progress

I don't use npm as heavily as I use git/docker so if you have an idea for satisfying a use case I'm not aware of, please raise an issue (and better yet a PR)

## Social

If you want to see what I (Jesse) am up to in terms of development, follow me on
[twitter](https://twitter.com/DuffieldJesse) or watch me program on
[twitch](https://www.twitch.tv/jesseduffield).
