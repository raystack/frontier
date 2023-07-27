# Admin Portal

The Admin Portal provides the Frontier administrators with a centralized interface for managing the Raystack/Frontier platform. This README will guide you through the installation, setup, and usage of the Admin Portal.

### Features

_Many of these features are still in development and represents an exhautive list of all the details in our Roadmap for the Admin Portal_

- Dashboard for an overview of system health and key metrics
- Tenant (organization) management for multi-tenancy support
- User authentication and access control
- User management with role-based access control
- Roles and permission management for fine-tuning Frontier's behavior
- Integration with external authentication identity providers/ configure Magic Links/ OTP based sign-ups
- Domain Verification for Organizations
- Audit Logs for monitoring and much more...

### Starting the Admin Portal

> Make sure you have the Frontier server up and running. For details refer [installations](./installation.md) and [configurations](./configurations.md)

Change the current working directory to ui in Frontier

```bash
$ cd ui
```

Create a **.env** file or export **`SHILD_API_URL`** environment variable for communication with the Frontier server.

```bash title=.env
# provide the frontier server url
SHILD_API_URL=http://localhost:8000
```

Finally, to start the Admin portal, run the following commands:

```bash
# install all dependencies
$ npm install
# start the Admin Portal UI
$ npm run dev
```

Open [http://localhost:5173/console](http://localhost:5173/console) with your browser to see the result.
