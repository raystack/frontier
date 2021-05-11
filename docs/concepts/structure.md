# Code Structure

This document describes the code structure of Shield, which will give a contributor an idea of where to make code changes to fix bugs or implement new features.

### server.ts

The server is initialized here.

### app

This folder contains the entry point of handling the native endpoints of Shield.
Any sub-folder inside `app` will have the entry points to handle the APIs. For eg: `/app/group` will contain the entry point of the code required to handle `/api/groups` endpoints.

### app/{domain}

As described earlier, any sub-folder inside `app` folder contains the entry point of the code required to handle various endpoints.

### app/{domain}/index.ts

This file contains all the routes for a specific domain. For eg: `app/group/index.ts` will container all the routes with `/api/groups` as base.

### app/{domain}/schema.ts

This file contains all the schema definitions used for validating requests and responses. The schemas have been defined using [Joi](https://github.com/sideway/joi)

### app/{domain}/handler.ts

The handler acts as a controller that validates and calls the appropriate resource method to handle the request.

### app/{domain}/resource.ts

This file contains methods with business logic to handle a request. These methods might interact with the database models to fetch data, or can even call other methods present in utils, other domains, lib, etc.

### config/composer.ts

The composer file registers all the plugins used by Shield, loads the configs, and exposes the method to construct the server.

### config/config.js

This file contains all the configs such as port, environment, etc.

### docs

This folder contains all the documentation deployed in Gitbook.

### factory

This folder contains all the factory files for models used for testing with mock database data.

### lib

The code placed in this folder can be extracted into an npm module itself, or they are wrappers of some npm module configured for Shield.

### lib/casbin

This folder contains code that wraps the casbin module.

### lib/index.ts

This file contains the factory that instantiates casbin enforcer.

### lib/JsonFilteredEnforcer.ts

A custom enforcer that loads policy subsets based on request and authorizes the request by using the JsonRoleManager.

### lib/JsonRoleManager.ts

This file contains a custom JSON-based role manager that maintains the data structure for the loaded policy, and checks whether a matching policy is found.

### lib/parser

Contains all the logic related to parsing the proxy endpoints YAML file.

### migration

Contains DB migration files

### model

Contains Typeorm DB models

### plugin

This folder contains plugins that listen to server events, or that contain logic that might be common to the whole server such as connecting to a database.

### plugin/iap

This plugin is responsible to authenticate the user using various authentication strategies.

### plugin/iam

This plugin determines whether a request is authorized or not. It picks the relevant data from the request and passes it to the casbin module.

### plugin/proxy

This plugin adds all the proxy routes configured by the user after parising it using `lib/parser`.

### test

This folder contains all the test files.

Architecture Invariant: It follows the same structure as the app structure. For eg: tests for all the `lib` files will go into `test/lib`

Architecture Invariant: We use integration tests for testing DB queries. So if you need to test some complex DB query made in a `resource.ts` file, you would have to setup a scenario and then test it

Architecture Invariant: The Casbin testing module only tests all the relevant scenarios of Shield's authorization model.

### types

Contains all the type definitions for modules that might not be compatible with typescript

### utils

Contains all the util functions.
