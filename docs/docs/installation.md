# Installation

We provide pre-built [binaries](https://github.com/odpf/shield/releases), [Docker Images](https://hub.docker.com/r/odpf/shield) and [Helm Charts](https://github.com/odpf/charts/tree/main/stable/shield)

## Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/shield/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

## Homebrew

```sh
# Install shield (requires homebrew installed)
$ brew install odpf/taps/shield

# Upgrade shield (requires homebrew installed)
$ brew upgrade shield

# Check for installed shield version
$ shield version
```

## Docker

### Prerequisites

- Docker installed

Run Docker Image

Shield provides Docker image as part of the release. Make sure you have Spicedb and postgres running on your local and run the following.

```sh
# Download docker image from docker hub
$ docker pull odpf/shield

# Run the following docker command with minimal config.
$ docker run -p 8080:8080 \
  -e SHIELD_DB_DRIVER=postgres \
  -e SHIELD_DB_URL=postgres://shield:@localhost:5432/shield?sslmode=disable \
  -e SHIELD_SPICEDB_HOST=spicedb.localhost:50051 \
  -e SHIELD_SPICEDB_PRE_SHARED_KEY=randomkey
  -v .config:.config
  odpf/shield serve
```

## Compiling from source

### Prerequisites

Shield requires the following dependencies:

- Golang (version 1.20 or above)
- Git

### Build

Run the following commands to compile `shield` from source

```shell
git clone git@github.com:odpf/shield.git
cd shield
make build
```

Use the following command to test

```shell
./shield version
```

Shield service can be started with the following command although there are few required [configurations](./reference/configurations.md) for it to start.

```sh
./shield server start
```
