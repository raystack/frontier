# Proto Definitions 

[ODPF/Proton](https://github.com/odpf/proton) is an open-source project developed by [ODPF](https://github.com/odpf) (Open DataOps Foundation) that provides a unified way to define and manage APIs in a microservices architecture. It aims to simplify the development and deployment of APIs by abstracting away the underlying implementation details.

In ODPF/Proton, the [Protobuf (protocol buffers)](https://protobuf.dev/) definitions are used to describe the structure and behavior of APIs. Protobuf is a language-agnostic binary serialization format developed by Google. It allows you to define the data models and API endpoints using a simple and concise syntax.

In the context of Shield, the protobuf definitions are used to define the API endpoints and associated access control policies. These definitions specify the request and response message structures, allowed methods, and any additional metadata required for authorization checks.

By leveraging the protobuf definitions with Protobuf compilers like protoc and buf, Shield automatically generates code for validating and enforcing the defined policies. It integrates with various frameworks and libraries to seamlessly enforce access control rules, ensuring that only authorized requests are allowed to access protected APIs and resources.

The current deployment uses the [v1beta1](https://github.com/odpf/proton/tree/main/odpf/shield/v1beta1) Shield API version.

:::info
While making any changes in Shield APIs, the makefile in Shield contains the Proton commit hash, which is utilized in Shield for generating protobuf files and documentation with `make proto` and `make doc` rules. 
:::

The **`make proto`** command creates [apidocs.swagger.yaml](https://github.com/odpf/shield/blob/main/proto/apidocs.swagger.json) specification which can be used to create a Postman collection to test these APIs. 

Besides this, one can import these files it in the [Swagger Editor](https://editor.swagger.io/) to visualize the Shield API documentation using the Swagger OpenAPI specification format.
