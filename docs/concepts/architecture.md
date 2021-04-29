# Architecture

## Technologies

Shield is developed with

- [node.js](https://nodejs.org/en/) - Javascript runtime
- [typescript](https://www.typescriptlang.org/) - Adds static type definitions to javascript
- [docker](https://www.docker.com/get-started) - container engine runs on top of operating system
- [hapi](https://hapi.dev/) - Web application framework
- [casbin](https://casbin.org/) - Access control library
- [typeorm](https://typeorm.io/#/) - Database agnostic sql query builder
- [postgres](https://www.postgresql.org/) - a relational database

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
