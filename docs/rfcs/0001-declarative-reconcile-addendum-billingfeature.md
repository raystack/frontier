# RFC 0001 addendum: the BillingFeature kind

| | |
|---|---|
| **Status** | Accepted. Implemented in the `reconcile` package and the `reconcile` and `export` commands. |
| **Author** | Rohil Surana |
| **Created** | 2026-11-07 |
| **Updated** | 2026-11-07 |
| **Extends** | [RFC 0001: Declarative management of platform resources](0001-declarative-reconcile.md) |

## Summary

This addendum adds one kind, BillingFeature, to the kinds in RFC 0001. It follows the five
rules of RFC 0001. RFC 0001 itself is not changed.

Before this kind, a file could only name a feature inside a BillingProduct. A feature's title
and metadata could not be set from a file. They lived only in the database. BillingFeature lets
a file set them, so a UI can show a feature's title and a tooltip from its metadata.

## Where it fits

| Kind | Object or value | Identity | A missing entry | How to remove |
|---|---|---|---|---|
| BillingFeature | object | name | plan fails | cannot be removed |

BillingFeature is an object, like BillingProduct and BillingPlan. Every in-scope feature on the
server must be in the file, and a missing one fails the plan. Features with names shorter than
three characters are out of scope (see below): they are not required, and listing one fails.

A feature cannot be deleted or retired. There is no delete API and no inactive state. So
`delete: true` is rejected, and a feature must stay in the file for as long as it exists on the
server. `frontier export billingfeature` lists every in-scope feature, so a file can start from it.

## The kind

An entry is a feature `{name, title, metadata}`.

- **Name.** The name is the identity. It is matched exactly as the server stores it, including
  case and spaces, the same way the API does. `Foo` and `foo` are two features, so a file name
  that differs from a server feature only by case is created as a new feature. A name made only
  of spaces is rejected.
- **Title and metadata.** Both follow rule 2. An omitted title is empty. Omitted metadata is
  `{}`. Metadata is one whole map: a file that drops a key removes it.
- **Product links.** Product links are owned by the BillingProduct kind. An update keeps the
  current links, and export never writes them. A new feature is created with no links.
- **Order.** In one file, BillingFeature must come before BillingProduct, because a product
  lists features by name.
- **Out of scope.** A feature with a name shorter than three characters cannot be sent through
  the API, so it is left out of the plan and the export.

## Server-side change

Rule 2 needs every default to be reachable through the API. Before this change, `UpdateFeature`
kept the old title and metadata when the request sent them empty. Now `UpdateFeature` writes the
title, metadata and product links exactly as sent, and finds the feature by name. Other paths
that save features, like creating a product or a plan, keep their old merge behavior.

The CLI and the server are expected to run the same version. An older server keeps the old
values when a file clears them, and the plan keeps showing the same update.
