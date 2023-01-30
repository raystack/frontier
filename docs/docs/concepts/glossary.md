# Glossary

## Backend

External Service which wants to use Shield for authz. It can verify access via Shield Proxy & API.

## Permission

Ability to carry out any action, in Shield or configured Backends.

## Principal

To whom we can grant Permission to. They can be of types:

1. User: A person or service account who can be a Principal. It is identified by Email ID.
2. Group: Collection of Users.
3. All Registered Users: Collection of users who have registered in Shield. Any user who registers in Shield becomes part of this Principal.


## Resource

Entity which needs authorization to be accessed. For example, a GCE instance is a resource over which we need permission such as edit & view.

## Resource Type

Classification that contains Resource instances. For example, GCE can be a resource type for GCE instances.

## Project

By which we can group Resources, of various different Resource Types, who have common environment.

## Organization

Organization is the root node in the hierarchy of Resources, being a collection of Projects.

## Namespace

Type of objects over which we want authorization. They are of two types:

1. System Namespace: Objects like Organization, Project & Team, over which we need authorization to actions such as adding user to team, adding user as owner of project.

2. Resource Namespace: Resources Types over which we need authorization. For example, we need edit & view permissions over GCE Instances.

## Role

Its an IAM Identity that describes what are the permissions one Principal has.

## Policy

Defines what Permission does a Role have.

## Entity

Instance of a namespace.

## Spicedb

[SpiceDB](https://github.com/authzed/spicedb) is a Zanzibar-inspired open source database system for managing security-critical application permissions.
