## Running locally

<details>
  <summary>Dependencies:</summary>

    - Git
    - Go 1.20 or above
    - PostgreSQL 13.2 or above

</details>

Clone the repo

```
git clone git@github.com:raystack/shield.git
```

Install all the golang dependencies

```
make install
```

Optional: build shield admin ui

```
make ui
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
