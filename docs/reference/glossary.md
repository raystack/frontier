# Glossary

This page contains all the definitions that you might need to understand Shield's usage.

* **Users**: Users of your application.
* **Groups**: Groups are just collections of users. Groups can also be used as attributes\(see below\) to map with resources.
* **Actions**: Actions are representation of your endpoints. For example: action name for `POST /books` would be `book.create`, similarly for `GET /books/:id` would be `book.read`, etc.
* **Roles**: Roles are nothing but collections of actions. Example: someone with a `Viewer` role can only perform `book.read`, whereas a `Manager` can perform both `book.read` and `book.create`.
* **Resources**: Resources are the assets you would like to protect via authorization. Example: All the individual books and e-books of a library are its resources, such as `The Theory of Relativity` book is one resource.
* **Attributes**: Resources get mapped with attributes that can be turn used while creating policies. Policies can get defined on attributes rather than individual resources to create better access management.
* **Permissions**: An array of permissions that should resolve to true for the user to be able to access your `uri`. You use it while configuring your endpoints in Shield.

