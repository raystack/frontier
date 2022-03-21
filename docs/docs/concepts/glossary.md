# Glossary

## User

A User represents a profile of a user. User has to create a profile with their name, email, and metadata.

## Group

A Group represents a group of [users](#user). There are predefined [roles](#role) for a user like Team Admin and Team Members. By default, every user gets a Team member role. Team admin can add and remove members from the Group.

## Role

Within a [namespace](#namespace), a role will get assigned to [users](#user). A role is needed to define [policy](#policy). There are predefined roles in a namespace. Users can also create custom roles.

## Namespace

Namespace provides scope for the [policies](#policy). There are predefined namespaces like [organization](#orgnaization), [project](#project), [group](#group), etc. Users can also create custom namespaces.

## Policy

A Policy defines what [actions](#action) a [role](#role) can perform in a [namespace](#namespace). All policies in a namespace will be used to generate schema for [spicedb](#spicedb).

## Action

Within a [namespace](#namespace), a role can perform certain actions. A action is needed to define [policy](#policy).

## Project

A Project is a scope in which a User can create and manage Resources.

## Orgnaization

An Orgnaization is just a group of Projects. [Groups](#group) also belongs to Orgnaization.

## Resource

Resources are just custom namespaces in which user can define policies. All Resources also gets default policies.

## Spicedb

[SpiceDB](https://github.com/authzed/spicedb) is a Zanzibar-inspired open source database system for managing security-critical application permissions.
