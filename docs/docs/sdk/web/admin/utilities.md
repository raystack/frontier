---
pagination_prev: sdk/web/admin/components
---

# Utilities

The SDK exports utility functions for working with ConnectRPC pagination and query transformation.

```tsx
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
  transformDataTableQueryToRQLRequest,
} from "@raystack/frontier/admin";
```

---

## `DEFAULT_PAGE_SIZE`

Default page size used for paginated queries.

```tsx
const DEFAULT_PAGE_SIZE = 50;
```

---

## `getConnectNextPageParam`

Returns the next page parameters for infinite queries using ConnectRPC pagination. Returns `undefined` when there are no more pages.

### Signature

```tsx
function getConnectNextPageParam<T extends ConnectRPCPaginatedResponse>(
  lastPage: T,
  queryParams: { query: RQLRequest },
  itemsKey?: string, // default: "organizations"
): { query: RQLRequest } | undefined;
```

### Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `lastPage` | `ConnectRPCPaginatedResponse` | The last page response from the API. |
| `queryParams` | `{ query: RQLRequest }` | The current query parameters. |
| `itemsKey` | `string` | The key in the response that contains the items array. Defaults to `"organizations"`. |

### Example

```tsx
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { getConnectNextPageParam } from "@raystack/frontier/admin";

const { data, fetchNextPage } = useInfiniteQuery(
  AdminServiceQueries.searchUsers,
  { query },
  {
    pageParamKey: "query",
    getNextPageParam: (lastPage) =>
      getConnectNextPageParam(lastPage, { query }, "users"),
  },
);
```

---

## `getGroupCountMapFromFirstPage`

Extracts group count data from the first page of an infinite query response. Useful for displaying group counts in `DataTable` column headers.

### Signature

```tsx
function getGroupCountMapFromFirstPage(
  infiniteData: { pages: ConnectRPCPaginatedResponse[] },
): Record<string, Record<string, number>>;
```

### Example

```tsx
const groupCountMap = infiniteData
  ? getGroupCountMapFromFirstPage(infiniteData)
  : {};

const columns = getColumns({ groupCountMap });
```

---

## `transformDataTableQueryToRQLRequest`

Converts an Apsara `DataTableQuery` object into an `RQLRequest` compatible with ConnectRPC APIs. Handles filters, sorting, pagination, and search.

### Signature

```tsx
function transformDataTableQueryToRQLRequest(
  query: DataTableQuery,
  options?: TransformOptions,
): RQLRequest;
```

### Options

| Option | Type | Description |
| --- | --- | --- |
| `defaultLimit` | `number` | Default page size. Defaults to `50`. |
| `fieldNameMapping` | `Record<string, string>` | Maps frontend field names to API field names (e.g. `{ createdAt: "created_at" }`). |

### Example

```tsx
import { transformDataTableQueryToRQLRequest } from "@raystack/frontier/admin";

const query = transformDataTableQueryToRQLRequest(tableQuery, {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
  },
});
```

---

## `useTerminology`

Hook that returns customizable entity labels based on the terminology config provided via `AdminConfigProvider`. Allows deployments to rename entities (e.g. "Organization" to "Workspace") across the UI without code changes.

The terminology is configured at runtime via the `/configs` endpoint:

```json
{
  "terminology": {
    "organization": { "singular": "Workspace", "plural": "Workspaces" },
    "user": { "singular": "Person", "plural": "People" }
  }
}
```

### Usage

```tsx
import { useTerminology } from "@raystack/frontier/admin";

const MyView = () => {
  const t = useTerminology();

  return (
    <div>
      <h1>{t.organization({ plural: true, case: "capital" })}</h1>
      <p>Create a new {t.organization({ case: "lower" })}</p>
    </div>
  );
};
```

### Options

Each entity function accepts an optional options object:

| Option | Type | Description |
| --- | --- | --- |
| `plural` | `boolean` | Use plural form. Defaults to `false`. |
| `case` | `"lower" \| "upper" \| "capital"` | Text casing. Returns the configured value as-is when omitted. |

### Available entities

`t.organization()`, `t.project()`, `t.team()`, `t.member()`, `t.user()`, `t.appName()`

---

## `useAdminPaths`

Hook that returns URL-safe slugs derived from terminology config, for use in dynamic routing.

```tsx
import { useAdminPaths } from "@raystack/frontier/admin";

const paths = useAdminPaths();
// Default: { organizations: "organizations", users: "users", projects: "projects", ... }
// With custom terminology: { organizations: "workspaces", users: "people", ... }

navigate(`/${paths.organizations}/${id}`);
```

---

## `ConnectRPCPaginatedResponse`

Type for paginated API responses.

```tsx
type ConnectRPCPaginatedResponse = {
  pagination?: RQLQueryPaginationResponse;
  group?: RQLQueryGroupResponse;
  [key: string]: unknown;
};
```

## `TransformOptions`

Options for `transformDataTableQueryToRQLRequest`.

```tsx
interface TransformOptions {
  defaultLimit?: number;
  fieldNameMapping?: Record<string, string>;
}
```
