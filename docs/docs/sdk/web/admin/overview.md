# Overview

`@raystack/frontier/admin` is a React component library that provides pre-built admin views for managing Frontier resources such as users, organizations, plans, webhooks, and more. Each view handles data fetching, pagination, and rendering internally using [ConnectRPC](https://connectrpc.com/) and [`@raystack/apsara`](https://github.com/raystack/apsara) UI components.

## Installation

```bash
npm install @raystack/frontier
```

## Prerequisites

### Peer Dependencies

The following packages must be installed in your application:

| Package | Version |
| --- | --- |
| `react` | `^18.2.0` |
| `@raystack/apsara` | `>=0.30.0` |

## Routing Pattern

The SDK is **router-agnostic**. It does not depend on `react-router-dom` or any other routing library. Navigation is handled via callback props (`onNavigateToUser`, `onSelectRole`, `onCloseDetail`, etc.), giving the consuming application full control over routing.

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
