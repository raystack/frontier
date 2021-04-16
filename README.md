# Shield

![build workflow](https://github.com/odpf/shield/actions/workflows/test.yml/badge.svg)
![package workflow](https://github.com/odpf/shield/actions/workflows/release.yml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/shield?logo=semantic-release)](Version)
[![Code Style](https://img.shields.io/badge/code_style-prettier-ff69b4.svg?style=flat-square)](https://prettier.io/)

Shield is a cloud native role-based authorization aware reverse-proxy service. With Shield, you can assign roles to users or groups of users to configure policies that determine whether a particular user has the ability to perform a certain action on a given resource.

## Key Features
Discover why users choose Shield as their authorization proxy

- **Policy Management**: Policies help you assign various roles to users/groups that determine their access to various resources
- **Group Management**: Group is nothing but another word for team. Shield provides APIs to add/remove users to/from a group, fetch list of users in a group along with their roles in the group, and fetch list of groups a user is part of.
- **Activity Logs**: Shield has APIs that store and retrieve all the access related logs. You can see who added/removed a user to/from group in these logs.
- **Reverse Proxy**: In addition to configuring access management, you can also use Shield as a reverse proxy to directly protect your endpoints by configuring appropriate permissions for them.
- **Google IAP**: Shield also utilizes Google IAP as an authentication mechanism. So if your services are behind a Google IAP, Shield will seemlessly integrate with it.
- **Runtime**: Shield can run inside containers or VMs in a fully managed runtime environment like Kubernetes. Shield also depends on a Postgres server to store data.

## How can I get started?

- [Guides](guides/overview.md) provide guidance on how to use Shield and configure it to your needs
- [Concepts](concepts/casbin.md) descibe the primary concepts and architecture behind Shield
- [Reference](reference/api.md) contains the list of all the APIs that Shield exposes
- [Contributing](contribute/contribution.md) contains resources for anyone who wants to contribute to Shield


## Technologies

Shield is developed with

- [node.js](https://nodejs.org/en/) - Javascript runtime
- [docker](https://www.docker.com/get-started) - container engine runs on top of operating system
- [hapi](https://hapi.dev/) - Web application framework
- [casbin](https://casbin.org/) - Access control library
- [typeorm](https://typeorm.io/#/) - Database agnostic sql query builder

## Running locally

In order to install this project locally, you can follow the instructions below:

```shell

$ git clone git@github.com:odpf/shield.git
$ cd shield
$ npm install
$ docker-compose up
```

Please refer [guides](guides/usage-reverse-proxy.md) section to know more about proxies.

If application is running successfully [click me](http://localhost:5000/ping) will open success message on a browser.

**Note** - before `docker-compose up` command run `docker` daemon locally.

Once running, you can find the Shield API documentation [on this link](http://localhost:5000/documentation)

## Contribute

Development of Shield happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and improvements. Read below to learn how you can take part in improving Shield.

Read our [contributing guide](docs/contribute/contribution.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Shield.

To help you get your feet wet and get you familiar with our contribution process, we have a list of [good first issues](https://github.com/odpf/shield/labels/good%20first%20issue) that contain bugs which have a relatively limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/shield/graphs/contributors).


## License

Shield is [Apache Licensed](LICENSE)
