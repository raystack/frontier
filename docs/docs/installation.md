# Installation

There are several approaches to install Frontier.

1. [Using a pre-compiled binary](#binary-cross-platform)
2. [Using the Docker image](#docker)
3. [Installing from source](#building-from-source)
4. [Helm Charts](https://github.com/raystack/charts/tree/main/stable/frontier)

### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/raystack/frontier/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`Frontier` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/raystack/frontier/releases) page:

```sh
$ brew install raystack/taps/frontier
```

To upgrade to the latest version:

```sh
$ brew upgrade frontier
```

#### Linux

`Frontier` is available as downloadable binaries from the [releases](https://github.com/raystack/frontier/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`Frontier` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/raystack/frontier/releases/latest) page:

```sh
scoop bucket add frontier https://github.com/raystack/scoop-bucket.git
```

To upgrade to the latest version:

```sh
scoop update frontier
```

### Docker

We provide ready to use Docker container images. To pull the latest image: Make sure you have Spicedb and postgres running on your local and run the following.

```sh
$ docker pull raystack/frontier
```

To pull a specific version:

```sh
$ docker pull raystack/frontier:v0.6.2-arm64
```

To run the docker image with minimum configurations:

```sh
$ docker run -p 8080:8080 \
  -e SHIELD_DB_DRIVER=postgres \
  -e SHIELD_DB_URL=postgres://frontier:@localhost:5432/frontier?sslmode=disable \
  -e SHIELD_SPICEDB_HOST=spicedb.localhost:50051 \
  -e SHIELD_SPICEDB_PRE_SHARED_KEY=randomkey
  -v .config:.config
  raystack/frontier server start
```

### Building from source

Begin by cloning this repository then you have two ways in which you can build Frontier:

- As a native executable
- As a docker image

Run either of the following commands to clone and compile Frontier from source

```bash
$ git clone git@github.com:raystack/frontier.git                 # (Using SSH Protocol)
$ git clone https://github.com/raystack/frontier.git             # (Using HTTPS Protocol)
```

#### As a native executable

To build frontier as a native executable, run `make` inside the cloned repository.

```bash
$ make
```

This will create the **`frontier`** binary in the root directory. Initialise server and client config file. Customise the `config.yaml` file with your local configurations.

```bash
$ ./frontier config init
```

Run database migrations

```bash
$ ./frontier server migrate
```

Start Frontier server

```bash
$ ./frontier server start
```

#### As a Docker image

Building Frontier's Docker image is just a simple, just run docker build command and optionally name the image

```bash
$ docker build . -t frontier
```

### Verifying the installation​

To verify if Frontier is properly installed, run `frontier --help` on your system. You should see help output. If you are executing it from the command line, make sure it is on your PATH or you may get an error about Frontier not being found.

```bash
$ ./frontier --help
```

### What's next

- See the [Configurations](./configurations.md) page on how to setup Frontier server and client
- See the [CLI Reference](./reference/cli.md) for a complete list of commands and options.
