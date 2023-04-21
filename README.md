# Shield

![build workflow](https://github.com/odpf/shield/actions/workflows/test.yml/badge.svg)
![package workflow](https://github.com/odpf/shield/actions/workflows/release.yml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/shield?logo=semantic-release)](Version)
[![Coverage Status](https://coveralls.io/repos/github/odpf/shield/badge.svg?branch=main)](https://coveralls.io/github/odpf/shield?branch=main)

Shield is a cloud native role-based authorization aware server and reverse-proxy system. With Shield, you can assign roles to users or groups of users to configure policies that determine whether a particular user has the ability to perform a certain action on a given resource.

<p align="center"><img src="./docs/static/img/overview.svg" /></p>

## Key Features

Discover why users choose Shield as their authorization proxy

- **Policy Management**: Policies help you assign various roles to users/groups that determine their access to various resources
- **Group Management**: Group is nothing but another word for team. Shield provides APIs to add/remove users to/from a group, fetch list of users in a group along with their roles in the group, and fetch list of groups a user is part of.
- **Activity Logs**: Shield has APIs that store and retrieve all the access related logs. You can see who added/removed a user to/from group in these logs.
- **Reverse Proxy**: In addition to configuring access management, you can also use Shield as a reverse proxy to directly protect your endpoints by configuring appropriate permissions for them.
- **Google IAP**: Shield also utilizes Google IAP as an authentication mechanism. So if your services are behind a Google IAP, Shield will seemlessly integrate with it.
- **Runtime**: Shield can run inside containers or VMs in a fully managed runtime environment like Kubernetes. Shield also depends on a Postgres server to store data.

## How can I get started?

- [Guides](guides/overview.md) provide guidance on how to use Shield and configure it to your needs
- [Concepts](concepts/casbin.md) descibe the primary concepts and architecture behind Shield
- [Reference](reference/api.md) contains the list of all the APIs that Shield exposes
- [Contributing](contribute/contribution.md) contains resources for anyone who wants to contribute to Shield

## Installation

Install Shield on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. Refer this for [installations](https://odpf.github.io/shield/docs/installation).

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/shield/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`shield` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/odpf/shield/releases/latest) page:

```sh
brew install odpf/tap/shield
```

To upgrade to the latest version:

```
brew upgrade shield
```

Check for installed shield version

```sh
shield version
```

#### Linux

`shield` is available as downloadable binaries from the [releases](https://github.com/odpf/shield/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`shield` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/odpf/shield/releases/latest) page:

```
scoop bucket add shield https://github.com/odpf/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update shield
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull odpf/shield:latest
```

To pull a specific version:

```
docker pull odpf/shield:v0.3.2
```

If you like to have a shell alias that runs the latest version of shield from docker whenever you type `shield`:

```
mkdir -p $HOME/.config/odpf
alias shield="docker run -e HOME=/tmp -v $HOME/.config/odpf:/tmp/.config/odpf --user $(id -u):$(id -g) --rm -it -p 3306:3306/tcp odpf/shield:latest"
```

## Usage

Shield is purely API-driven. It is very easy to get started with Shield. It provides CLI, HTTP and GRPC APIs for simpler developer experience.

#### CLI

Shield CLI is fully featured and simple to use, even for those who have very limited experience working from the command line. Run `shield --help` to see list of all available commands and instructions to use.

List of commands

```
shield --help
```

Print command reference

```sh
shield reference
```

#### API

Shield provides a fully-featured GRPC and HTTP API to interact with Shield server. Both APIs adheres to a set of standards that are rigidly followed. Please refer to [proton](https://github.com/odpf/proton/tree/main/odpf/shield/v1beta1) for GRPC API definitions.

## Running locally

<details>
  <summary>Dependencies:</summary>

    - Git
    - Go 1.17 or above
    - PostgreSQL 13.2 or above

</details>

Clone the repo

```
git clone git@github.com:odpf/shield.git
```

Install all the golang dependencies

```
make install
```

Build shield binary file

```
make build
```

Init config

```
cp internal/server/config.yaml config.yaml
./shield config init
```

Run database migrations

```
./shield server migrate -c config.yaml
```

Start shield server

```
./shield server start -c config.yaml
```

## Running tests

```sh
# Running all unit tests
$ make test

# Print code coverage
$ make coverage
```

## Contribute

Development of Shield happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and
improvements. Read below to learn how you can take part in improving Shield.

Read our [contributing guide](https://odpf.github.io/shield/docs/contribute/contribution) to learn about our development process, how to propose
bugfixes and improvements, and how to build and test your changes to Shield.

To help you get your feet wet and get you familiar with our contribution process, we have a list of
[good first issues](https://github.com/odpf/shield/labels/good%20first%20issue) that contain bugs which have a relatively
limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/shield/graphs/contributors).

## License

Shield is [Apache 2.0](LICENSE) licensed.
