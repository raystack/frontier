# Shield

![build workflow](https://github.com/raystack/shield/actions/workflows/test.yml/badge.svg)
![package workflow](https://github.com/raystack/shield/actions/workflows/release.yml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/raystack/shield?logo=semantic-release)](Version)
[![Coverage Status](https://coveralls.io/repos/github/raystack/shield/badge.svg?branch=main)](https://coveralls.io/github/raystack/shield?branch=main)

Shield is an identity and access management tool designed to help organizations secure their systems and data. With Shield, you can manage user authentication and authorization across all your applications and services, ensuring that only authorized users have access to your valuable resources.

<p align="center"><img src="./docs/static/img/overview.svg" /></p>

## Key Features

Discover why users choose Shield as their authorization server

- **User management** Create and manage user accounts for all your applications and services.
- **Organization management** Manage multiple tenants, each with their own set of users, applications, and services.
- **Project management** Organize your resources into projects and manage access permissions for each project.
- **Group management** Create and manage groups of users with different access levels across projects and applications.
- **Authentication** Multiple authentication strategies like Email OTP, Social Login for human users and API keys, RSA JWT based for machine users.
- **Authorization** Role based access control with policies to bind a user to its access level.
- **Billing management** Manage billing and subscriptions for your users.
- **Audit** Audit all user activity and access related logs.
- **Reporting** Generate reports on user activity and access levels.

## How can I get started?

- [Introduction](docs/docs/introduction.md) provide guidance on how to use Shield and configure it to your needs
- [Concepts](docs/docs/concepts/architecture.md) descibe the primary concepts and architecture behind Shield
- [Reference](docs/docs/reference/api-definitions.md) contains the list of all the APIs that Shield exposes
- [Contributing](docs/docs/contribution/contribute.md) contains resources for anyone who wants to contribute to Shield

## Installation

Install Shield on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. Refer this for [installations](https://raystack.github.io/shield/docs/installation).

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/raystack/shield/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`shield` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/raystack/shield/releases/latest) page:

```sh
brew install raystack/tap/shield
```

To upgrade to the latest version:

```
brew upgrade shield
```

#### Linux

`shield` is available as downloadable binaries from the [releases](https://github.com/raystack/shield/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`shield` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/raystack/shield/releases/latest) page:

```
scoop bucket add shield https://github.com/raystack/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update shield
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull raystack/shield:latest
```

To pull a specific version:

```
docker pull raystack/shield:0.6.2
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

Shield provides a fully-featured GRPC and HTTP API to interact with Shield server. Both APIs adheres to a set of standards that are rigidly followed. Please refer to [proton](https://github.com/raystack/proton/tree/main/raystack/shield/v1beta1) for GRPC API definitions.

## Contribute

Development of Shield happens on GitHub, and we are grateful to the community for contributing bugfixes and
improvements.

Read our [contribution guide](https://raystack.github.io/shield/docs/contribute/contribution) to learn about our development process, how to propose
bugfixes and improvements, and how to build and test your changes to Shield.

To help you get your feet wet and get you familiar with our contribution process, we have a list of
[good first issues](https://github.com/raystack/shield/labels/good%20first%20issue) that contain bugs which have a relatively
limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/raystack/shield/graphs/contributors).

## License

Shield is [Apache 2.0](LICENSE) licensed.
