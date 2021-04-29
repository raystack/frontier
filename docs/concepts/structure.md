# Structure

## Handling Requests

Shield uses the following structure to handle a request.

### index.ts

Contains all the routes for a specific domain.

### schema.ts

Contains all the schema definitions that are used for validation requests and responses

### handler.ts

Acts as a controller that validates and also calls the appropriate methods to handle the request.

### resource.ts

Contains all the business logic to handle a request
