[![Go Reference](https://pkg.go.dev/badge/github.com/amatsagu/tempest.svg)](https://pkg.go.dev/github.com/amatsagu/tempest)
[![Go Report](https://goreportcard.com/badge/github.com/amatsagu/tempest)](https://goreportcard.com/report/github.com/amatsagu/tempest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/amatsagu/tempest)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/github/license/Amatsagu/tempest)](https://github.com/amatsagu/tempest/blob/development/LICENSE)
[![Maintenance Status](https://img.shields.io/maintenance/yes/2025)](https://github.com/amatsagu/tempest)
[![CodeQL](https://github.com/amatsagu/tempest/actions/workflows/github-code-scanning/codeql/badge.svg?branch=development)](https://github.com/amatsagu/tempest/actions/workflows/github-code-scanning/codeql)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)

<img align="left" src="/.github/tempest-logo.png" width=192 alt="Tempest library logo">

# Tempest
Tempest is a [Discord](https://discord.com) API wrapper for Applications, written in [Golang](https://golang.org/). It aims to be fast, use minimal caching and be easier to use than other Discord API wrappers using http.

It was created as a better alternative to [discord-interactions-go](https://github.com/bsdlp/discord-interactions-go) which is "low level" and outdated.

## Summary
1. [HTTP vs Gateway](#http-vs-gateway)
2. [Special features](#special-features)
3. [Getting Started](#getting-started)
4. [Troubleshooting](#troubleshooting)
5. [Contributing](#contributing)

### HTTP vs Gateway
**TL;DR**: you probably should be using libraries like [DiscordGo](https://github.com/bwmarrin/discordgo) unless you know why you're here.

There are two ways for bots to receive events from Discord. Most API wrappers such as **DiscordGo** use a WebSocket connection called a "gateway" to receive events, but **Tempest** receives interaction events over HTTP. Using http hooks lets you scale code more easily & reduce resource usage at cost of greatly reduced number of events you can use. You can easily create bots for roles, minigames, custom messages or admin utils but it'll be very difficult / impossible to create music or moderation bots.

### Special features
* [Easy to use & efficient handler for (/) commands & auto complete interactions](https://pkg.go.dev/github.com/amatsagu/tempest#Client.RegisterCommand)
    - Deep control with [command middleware(s)](https://pkg.go.dev/github.com/amatsagu/tempest#ClientOptions)
* [Exposed REST](https://pkg.go.dev/github.com/amatsagu/tempest#Client.Rest)
* [Easy component & modal handling](https://pkg.go.dev/github.com/amatsagu/tempest#Client.AwaitComponent)
    - Works with buttons, select menus, text inputs and modals,
    - Supports timeouts & gives a lot of freedom,
    - Works for both [static](https://pkg.go.dev/github.com/amatsagu/tempest#Client.RegisterComponent) and [dynamic](https://pkg.go.dev/github.com/amatsagu/tempest#Client.AwaitModal) ways
* [Simple way to sync (/) commands with API](https://pkg.go.dev/github.com/amatsagu/tempest#Client.SyncCommands)
* Request failure auto recovery (3 attempts by default)
    - On failed attempts *(probably due to internet connection)*, it'll try again set number of times before returning error
* No Discord data caching by default

### Getting started
1. Install with: `go get -u github.com/amatsagu/tempest`
2. Check [example](https://github.com/amatsagu/tempest/blob/master/example) with few simple commands.



## Troubleshooting
For help feel free to open an issue on github.
You can also inivite to contact me on [discord](https://discord.com/users/390394829789593601).

## Contributing
All contributions are welcomed.
Few rules before making a pull request:
* Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
* Add link to document for new structs
    - Since `v1.1.0`, all structs should have links to corresponding discord docs



[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FAmatsagu%2FTempest.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FAmatsagu%2FTempest?ref=badge_large)
