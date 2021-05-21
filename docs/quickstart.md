# Quickstart

We have created an example to help you get started with Shield. The library application mentioned [here](guides/managing_policies.md) has been implemented in the `example` directory.

### Get started in a few simple steps

```text
# clone the repository if you don't have it yet
git clone https://github.com/odpf/shield.git && cd shield

docker-compose -f example/docker-compose.yml up

```

### Output

Wait for some time until you see this output in the terminal.

```text
 SETTING UP LIBRARY DATAs

 admin@library.com was given super admin privileges

 Einstein was added
 Newton was added
 Darwin was added

 Scientists group was added
 Mathematicians group was added

 Team Admin role was created
 Book Manager role was created

 relativity-the-special-general-theory resource was mapped with {"category":"physics","group":"62ac285f-a98b-42bf-a119-b30bc8f5c22e"}

 Einstein was given Book Manager role in Scientists group
 Darwin was given Book Manager role in Scientists group
 Newton was given Team Admin role in Scientists group
```

```text

CHECKING ACCESS

Einstein has book.update action access to relativity-the-special-general-theory
Darwin does not have book.update action access to relativity-the-special-general-theory

```

You can check `http://localhost:5000/documentation` to view all the APIs of the Shield.

### Example Explanation

The Library Application example shows how to protect your application with Shield using the reverse proxy configuration.

- **example/library/index.js**: hosts contains the Library Application server
- **example/library/check.js**: tries to simulate requests from different users and outputs whether they were authorized
- **example/library/setup.js**: contains code to ingest data into Shield using its APIs
- **example/proxies**: contains the reverse proxy configuration that is needed by Shield
- **example/Dockerfile**: contains the docker config that you would need to build the Shield server
