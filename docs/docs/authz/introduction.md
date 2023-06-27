---
title: Overview
---

# Overview

Authorization is the process of determining whether a user is allowed to perform an action. In the context of
Shield, authorization is the process of determining whether a user is allowed to perform an action on a resource.
This is done after the system has already confirmed that user has proven their identity (authentication).

Shield authorization is based on the [Role Based Access Control (RBAC)](https://en.wikipedia.org/wiki/Role-based_access_control) model.
In RBAC, access is granted to users based on their roles. A role is a collection of permissions that can be assigned 
to a user. Permissions determine what actions are allowed on a resource. When a role is assigned to a user, the user 
is granted all the permissions that the role contains. 

