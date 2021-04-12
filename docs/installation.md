# Installation

Installing Shield on any system is straightforward. We provide a docker image, which you can pull from github packages.

## Pre-requisites

- To run shield on production, you would need to host your own `postgres` database.
- You also need to create a `.env` file by using `.env.sample` as a reference and set all the values.

## Usage with Docker

You can create the following `Dockerfile` to deploy Shield

```text
FROM docker.pkg.github.com/odpf/shield/shield:0.1.13
COPY proxies proxies
COPY .env .env
```

Please refer [guides](guides/usage-reverse-proxy.md) section to know more about proxies.
