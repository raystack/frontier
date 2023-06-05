# Installation

There are several approaches to install Shield.

1. [Using a pre-compiled binary](#binary-cross-platform)
2. [Using the Docker image](#docker)
3. [Installing from source](#building-from-source)
4. [Helm Charts](https://github.com/odpf/charts/tree/main/stable/shield)

### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/shield/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`Shield` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/odpf/shield/releases) page:

```sh
$ brew install odpf/taps/shield
```

To upgrade to the latest version:

```sh
$ brew upgrade shield
```

#### Linux

`Shield` is available as downloadable binaries from the [releases](https://github.com/odpf/shield/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`Shield` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/odpf/shield/releases/latest) page:

```sh
scoop bucket add shield https://github.com/odpf/scoop-bucket.git
```

To upgrade to the latest version:

```sh
scoop update shield
```

### Docker

We provide ready to use Docker container images. To pull the latest image: Make sure you have Spicedb and postgres running on your local and run the following.

```sh
$ docker pull odpf/shield
```

To pull a specific version:

```sh
$ docker pull odpf/shield:v0.5.1
```

To run the docker image with minimum configurations:

```sh
$ docker run -p 8080:8080 \
  -e SHIELD_DB_DRIVER=postgres \
  -e SHIELD_DB_URL=postgres://shield:@localhost:5432/shield?sslmode=disable \
  -e SHIELD_SPICEDB_HOST=spicedb.localhost:50051 \
  -e SHIELD_SPICEDB_PRE_SHARED_KEY=randomkey
  -v .config:.config
  odpf/shield server start
```

### Building from source

Begin by cloning this repository then you have two ways in which you can build Shield:

- As a native executable
- As a docker image

Run either of the following commands to clone and compile Shield from source

```bash
$ git clone git@github.com:odpf/shield.git                 # (Using SSH Protocol)
$ git clone https://github.com/odpf/shield.git             # (Using HTTPS Protocol)
```

#### As a native executable

To build shield as a native executable, run `make` inside the cloned repository.

```bash
$ make
```

This will create the **`shield`** binary in the root directory. Initialise server and client config file. Customise the `config.yaml` file with your local configurations.

```bash
$ ./shield config init
```

Run database migrations

```bash
$ ./shield server migrate
```

Start Shield server

```bash
$ ./shield server start
```

#### As a Docker image

Building Shield's Docker image is just a simple, just run docker build command and optionally name the image

```bash
$ docker build . -t shield
```

### Verifying the installation​

To verify if Shield is properly installed, run `shield --help` on your system. You should see help output. If you are executing it from the command line, make sure it is on your PATH or you may get an error about Shield not being found.

```bash
$ ./shield --help
```

### What's next

- See the [Configurations](./configurations.md) page on how to setup Shield server and client
- See the [CLI Reference](./reference/cli.md) for a complete list of commands and options.
