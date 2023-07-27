## Running locally

<details>
  <summary>Dependencies:</summary>

    - Git
    - Go 1.20 or above
    - PostgreSQL 13.2 or above

</details>

Clone the repo

```
git clone git@github.com:raystack/frontier.git
```

Install all the golang dependencies

```
make install
```

Optional: build frontier admin ui

```
make ui
```

Build frontier binary file

```
make build
```

Init config

```
cp internal/server/config.yaml config.yaml
./frontier config init
```

Run database migrations

```
./frontier server migrate -c config.yaml
```

Start frontier server

```
./frontier server start -c config.yaml
```

## Running tests

```sh
# Running all unit tests
$ make test

# Print code coverage
$ make coverage
```
