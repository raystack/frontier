# Deployment

Installing Shield on any system is straightforward. We provide a docker image, which you can pull from github packages.

## Pre-requisites

- To run Shield on production, you would need to host your own `postgres` database.
- You also need to create a `.env` file by using `.env.sample` as a reference and set all the values.

## Running with Docker

You can create the following `Dockerfile` to deploy Shield

```text
FROM docker.pkg.github.com/odpf/shield/shield:0.1.13
COPY proxies proxies
COPY .env .env
```

## Deploying with Helm

You can also use [Shield's helm chart](https://github.com/odpf/charts/tree/main/stable/shield) to deploy it on a K8 cluster.

## Usage Strategies

### Using Shield as a reverse proxy

As described [here](https://odpf.gitbook.io/shield/guides/usage_reverse_proxy), you can also use Shield as a reverse proxy, which either forwards or forbids requests based on authorization checks.

![](../.gitbook/assets/overview.svg)

### Using Shield as an external Authorization microservice

As described [here](https://odpf.gitbook.io/shield/guides/managing_policies#checking-access), you can use Shield as an external authorization microservice, which authorizes calls made from your service.

![](../.gitbook/assets/service.png)
