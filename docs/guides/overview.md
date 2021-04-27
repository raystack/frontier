# Overview

The following topics will describe how to use Shield.

## Using the check-access API

You can use Shield as an authorization microservice, which stores all the authorization-related policies, and exposes the check-access API, which you can call from within your server to check whether a user is authorized.

## Using as a reverse proxy

You can also use Shield as a reverse proxy by configuring all your routes with it. In this case, Shield will check whether a user has the necessary permissions before forwarding the request to your endpoint.

## Deploying Shield

This section contains guides, best practices, and advice related to deploying Shield in production.

{% page-ref page="deployment.md" %}

