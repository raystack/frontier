# Overview

The following topics will describe how to use Shield.

## Managing Policies

Check the example on this page to learn more about Shield's Policy Management APIs.

{% page-ref page="managing\_policies.md" %}

## Using as an authorization service

One way to protect your endpoints is to use Shield as an authorization microservice, which stores all the authorization-related policies, and exposes the check-access API, which you can call from within your server to check whether a user is authorized.

{% page-ref page="using\_auth\_server.md" %}

## Using as a reverse proxy

Another way to protect your endpoints is to use Shield as a reverse proxy by configuring all your routes with it. In this case, Shield will check whether a user has the necessary permissions before forwarding the request to your endpoint.

{% page-ref page="usage\_reverse\_proxy.md" %}

## Deploying Shield

This section contains guides, best practices, and advice related to deploying Shield in production.

{% page-ref page="deployment.md" %}

## Authentication

This section describes how Shield authenticates a request.

{% page-ref page="authentication.md" %}
