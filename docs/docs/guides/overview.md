# Overview

The following topics will describe how to use Shield. It respects multi-tenancy using namespace, which can either be a system namespace or a resource namespace.
System namespace is composed of a `backend` and a `resource type` which allows onboard multiple instances of a service by changing the backend. While resource namespaces allows to onboard multiple `organizations`, `project` and `groups`.

## Managing Organizations

Organizations are the top most level object in Sheild's system.

## Managing Projects

Project comes under an organization, and they can have multiple resources belonging to them.

## Managing Resources

Resources have some basic information about the resources being created on the backend. They can be used for authorization purpose.

## Managing Groups

Groups fall under anorganization too, an they are a colection of users with different roles.

## Managing Namespaces, Policies, Roles and Actions

All of these are managed by configuration files and shall not be modified via APIs.

## Managing Users and their Metadata

User represent a real life user distinguished by their emails.

## Managing Relations

Relations are a copy of the relationships being managed in SpiceDB.

## Checking Permission

Shield provides an API to check if a user has a  certain permissions on a resource.

## Where to go next?

We recomment you to check all the guides for having a clear understanding of the APIs. For testing these APIs on local, you can import the [Swagger](https://github.com/odpf/shield/blob/main/proto/apidocs.swagger.json).


