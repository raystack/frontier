# Shield

Shield is an authorization aware proxy service 

## Table of contents
* [Technologies](#technologies)
* [Getting Started](#gettingstarted)
* [Documentation](#documentation)
* [Contributing](#contributing)
    * [Using the issue tracker](#usingtheissuetracker)
    * [Changing the code-base](#changingthecodebase)
* [License](#license)    


## [Technologies](#technologies)
Shield is developed with 
* [node.js](https://nodejs.org/en/) - Javascript runtime 
* [docker](https://www.docker.com/get-started) - container engine runs on top of operating system   
* [@hapi](https://hapi.dev/) - Web application framework
* [casbin](https://casbin.org/) - Access control library
* [typeorm](https://typeorm.io/#/) - Database agnostic sql query builder

## [Getting Started](#gettingstarted)
In order to install this project locally, you can follow the instructions below:
 
```shell

$ git clone git@github.com:odpf/shield.git
$ cd shield
$ npm install
$ docker-compose up
```
If application is running successfully [click me](http://localhost:5000/ping) will open success message on a browser.

**Note** - before `docker-compose up` command run `docker` daemon locally.

## [Documentation](#documentation)
You can find the Shield API documentation [on this link](http://localhost:5000/documentation)

## [Contributing](#contributing)
Contribute to our source code and to make Shield even better. Here are the [contributing guidelines]() we'd like you to follow:

### [Using the issue tracker](#usingtheissuetracker)

Use the issue tracker to suggest feature requests, report bugs, and ask questions.
This is also a great way to connect with the developers of the project as well
as others who are interested in this solution.

Use the issue tracker to find ways to contribute. Find a bug or a feature, mention in
the issue that you will take on that effort, then follow the _Changing the code-base_
guidance below.

### [Changing the code-base](#changingthecodebase)

Generally speaking, you should fork this repository, make changes in your
own fork, and then submit a pull request. All new code should have associated
unit tests that validate implemented features and the presence or lack of defects.
Additionally, the code should follow any stylistic and architectural guidelines
prescribed by the project. In the absence of such guidelines, mimic the styles
and patterns in the existing code-base.
## [License](#license)
Shield is [Apache Licensed](LICENSE)