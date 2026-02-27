# Components

All view components are exported from `@raystack/frontier/admin`.

```tsx
import { UsersView, PlansView, OrganizationList } from "@raystack/frontier/admin";
```

---

## UsersView

Displays a paginated, searchable list of users. When a `selectedUserId` is provided, renders the user detail view with a security panel.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedUserId` | `string` | No | When set, shows the detail view for this user. |
| `onCloseDetail` | `() => void` | No | Called when the detail view is closed. |
| `onExportUsers` | `() => Promise<void>` | No | Called when the user clicks the export button. |
| `onNavigateToUser` | `(userId: string) => void` | No | Called when a user row is clicked. |

### Example

```tsx
import { UsersView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function UsersPage() {
  const { userId } = useParams();
  const navigate = useNavigate();

  return (
    <UsersView
      selectedUserId={userId}
      onNavigateToUser={(id) => navigate(`/users/${id}/security`)}
      onCloseDetail={() => navigate("/users")}
    />
  );
}
```

---

## OrganizationList

Displays a paginated, searchable list of organizations with filtering and grouping support.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `appName` | `string` | No | Application name shown in the page title. |
| `onNavigateToOrg` | `(id: string) => void` | No | Called when an organization row is clicked. |
| `onExportCsv` | `() => Promise<void>` | No | Called when the export button is clicked. |
| `organizationTypes` | `string[]` | No | List of organization types for filtering. |
| `appUrl` | `string` | No | Base URL of the application. |
| `countries` | `string[]` | No | List of countries for filtering. |

### Example

```tsx
import { OrganizationList } from "@raystack/frontier/admin";
import { useNavigate } from "react-router-dom";

function OrganizationsPage() {
  const navigate = useNavigate();

  return (
    <OrganizationList
      onNavigateToOrg={(id) => navigate(`/organizations/${id}/members`)}
    />
  );
}
```

---

## OrganizationDetails

Renders the detail layout for a single organization including a navbar with tabs (members, security, projects, etc.). The consuming app must provide child routes via the `children` prop.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `organizationId` | `string | undefined` | Yes | The ID of the organization to display. |
| `currentPath` | `string` | Yes | Current URL pathname, used for active tab highlighting. |
| `onNavigate` | `(path: string) => void` | Yes | Called when a nav tab is clicked. |
| `children` | `React.ReactNode` | No | Child content (e.g. `<Outlet />` from your router). |
| `onExportMembers` | `() => Promise<void>` | No | Called when exporting members. |
| `onExportProjects` | `() => Promise<void>` | No | Called when exporting projects. |
| `onExportTokens` | `() => Promise<void>` | No | Called when exporting tokens. |
| `appUrl` | `string` | No | Base URL of the application. |
| `tokenProductId` | `string` | No | Product ID for token management. |
| `countries` | `string[]` | No | List of countries for filtering. |
| `organizationTypes` | `string[]` | No | List of organization types. |

### Sub-page Components

These components are meant to be rendered as children of `OrganizationDetails` via your router:

- `OrganizationMembersPage`
- `OrganizationSecurity`
- `OrganizationProjectssPage`
- `OrganizationInvoicesPage`
- `OrganizationTokensPage`
- `OrganizationApisPage`

### Example

```tsx
import {
  OrganizationDetails,
  OrganizationMembersPage,
  OrganizationSecurity,
} from "@raystack/frontier/admin";
import { useParams, useNavigate, useLocation, Outlet } from "react-router-dom";

function OrgDetailsPage() {
  const { orgId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <OrganizationDetails
      organizationId={orgId}
      currentPath={location.pathname}
      onNavigate={navigate}
    >
      <Outlet />
    </OrganizationDetails>
  );
}
```

---

## PlansView

Displays a list of billing plans. When a plan is selected, opens a side sheet with plan details.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedPlanId` | `string` | No | When set, opens the detail sheet for this plan. |
| `onCloseDetail` | `() => void` | No | Called when the detail sheet is closed. |
| `appName` | `string` | No | Application name shown in the page title. |

### Example

```tsx
import { PlansView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function PlansPage() {
  const { planId } = useParams();
  const navigate = useNavigate();

  return (
    <PlansView
      selectedPlanId={planId}
      onCloseDetail={() => navigate("/plans")}
    />
  );
}
```

---

## WebhooksView

Displays a list of webhooks with create/update side sheets.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedWebhookId` | `string` | No | When set, opens the update sheet for this webhook. |
| `createOpen` | `boolean` | No | When `true`, opens the create webhook sheet. |
| `onCloseDetail` | `() => void` | No | Called when a sheet is closed. |
| `onSelectWebhook` | `(id: string) => void` | No | Called when a webhook row is clicked. |
| `onOpenCreate` | `() => void` | No | Called when the "New Webhook" button is clicked. |
| `enableDelete` | `boolean` | No | When `true`, enables webhook deletion. |

### Example

```tsx
import { WebhooksView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function WebhooksPage() {
  const { webhookId, action } = useParams();
  const navigate = useNavigate();

  return (
    <WebhooksView
      selectedWebhookId={webhookId}
      createOpen={action === "create"}
      onCloseDetail={() => navigate("/webhooks")}
      onSelectWebhook={(id) => navigate(`/webhooks/${id}`)}
      onOpenCreate={() => navigate("/webhooks/create")}
    />
  );
}
```

---

## PreferencesView

Displays platform preferences with inline editing. When a preference is selected, shows a detail editor.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedPreferenceName` | `string` | No | When set, opens the detail panel for this preference. |
| `onCloseDetail` | `() => void` | No | Called when the detail panel is closed. |

### Example

```tsx
import { PreferencesView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function PreferencesPage() {
  const { name } = useParams();
  const navigate = useNavigate();

  return (
    <PreferencesView
      selectedPreferenceName={name}
      onCloseDetail={() => navigate("/preferences")}
    />
  );
}
```

---

## RolesView

Displays a list of platform roles. When a role is selected, opens a side sheet with role details.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedRoleId` | `string` | No | When set, opens the detail sheet for this role. |
| `onSelectRole` | `(roleId: string) => void` | No | Called when a role row is clicked. |
| `onCloseDetail` | `() => void` | No | Called when the detail sheet is closed. |
| `appName` | `string` | No | Application name shown in the page title. |

### Example

```tsx
import { RolesView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function RolesPage() {
  const { roleId } = useParams();
  const navigate = useNavigate();

  return (
    <RolesView
      selectedRoleId={roleId}
      onCloseDetail={() => navigate("/roles")}
      onSelectRole={(id) => navigate(`/roles/${id}`)}
    />
  );
}
```

---

## AdminsView

Displays a list of platform super admins (users and service users).

This component has no props.

### Example

```tsx
import { AdminsView } from "@raystack/frontier/admin";

function AdminsPage() {
  return <AdminsView />;
}
```

---

## AuditLogsView

Displays a paginated, searchable list of audit log records with filtering and grouping.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `appName` | `string` | No | Application name shown in the page title. |
| `onExportCsv` | `(query: RQLRequest) => Promise<void>` | No | Called when the export button is clicked. Receives the current query. |
| `onNavigate` | `(path: string) => void` | No | Called for navigation actions within audit logs. |

### Example

```tsx
import { AuditLogsView } from "@raystack/frontier/admin";

function AuditLogsPage() {
  return <AuditLogsView appName="MyApp" />;
}
```

---

## InvoicesView

Displays a paginated list of invoices.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `appName` | `string` | No | Application name shown in the page title. |

### Example

```tsx
import { InvoicesView } from "@raystack/frontier/admin";

function InvoicesPage() {
  return <InvoicesView />;
}
```

---

## ProductsView

Displays a list of billing products. When a product is selected, opens a detail sheet.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `selectedProductId` | `string` | No | When set, opens the detail sheet for this product. |
| `onSelectProduct` | `(productId: string) => void` | No | Called when a product row is clicked. |
| `onCloseDetail` | `() => void` | No | Called when the detail sheet is closed. |
| `onNavigateToPrices` | `(productId: string) => void` | No | Called when navigating to the prices view for a product. |
| `appName` | `string` | No | Application name shown in the page title. |

### Example

```tsx
import { ProductsView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

function ProductsPage() {
  const { productId } = useParams();
  const navigate = useNavigate();

  return (
    <ProductsView
      selectedProductId={productId}
      onCloseDetail={() => navigate("/products")}
      onSelectProduct={(id) => navigate(`/products/${id}`)}
      onNavigateToPrices={(id) => navigate(`/products/${id}/prices`)}
    />
  );
}
```

---

## ProductPricesView

Displays the list of prices for a specific product.

### Props

| Prop | Type | Required | Description |
| --- | --- | --- | --- |
| `productId` | `string` | Yes | The product ID to show prices for. |
| `appName` | `string` | No | Application name shown in the page title. |
| `onBreadcrumbClick` | `(item: { name: string; href?: string }) => void` | No | Called when a breadcrumb item is clicked. |

### Example

```tsx
import { ProductPricesView } from "@raystack/frontier/admin";
import { useParams, useNavigate } from "react-router-dom";

function ProductPricesPage() {
  const { productId } = useParams();
  const navigate = useNavigate();

  return (
    <ProductPricesView
      productId={productId!}
      onBreadcrumbClick={(item) => item.href && navigate(item.href)}
    />
  );
}
```
