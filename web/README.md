# Frontier Web

Frontend monorepo for [Frontier](https://github.com/raystack/frontier). Managed with [pnpm](https://pnpm.io/) workspaces and [Turborepo](https://turbo.build/).

## What's inside

| Package            | Description                                                              |
| ------------------ | ------------------------------------------------------------------------ |
| `apps/admin`       | Admin dashboard (Vite + React 19, ConnectRPC, Apsara). Embedded into the Frontier server binary via `embed.go`. |
| `apps/client-demo` | Reference app demonstrating the `@raystack/frontier` SDK (Vite + React 19). |
| `sdk`              | `@raystack/frontier` — the JS/React SDK for auth and account components.  |
| `tools/*`          | Shared `eslint-config` and `tsconfig` presets.                           |

## Prerequisites

- [Node.js](https://nodejs.org/) `>= 22`
- [pnpm](https://pnpm.io/) `>= 10.19.0`

## Getting started

```sh
# from frontier/web
pnpm install
```

### Run a dev server

Run everything in parallel:

```sh
pnpm dev
```

Or target a single app with a Turbo filter:

```sh
pnpm dev --filter=admin
pnpm dev --filter=client-demo
```

Or run each package directly from its own directory. Build the SDK first, then
start whichever app you're working on:

```sh
cd web/sdk
pnpm build

cd web/apps/admin
pnpm dev

cd web/apps/client-demo
pnpm dev
```

`client-demo` runs at [http://localhost:3000](http://localhost:3000) and proxies API
requests to a running Frontier server (see [Configure the backend](#configure-the-backend)).

### Configure the backend

Each app talks to a running Frontier instance via a `.env` file at its own root.
These files are git-ignored, so create them before starting a dev server. For a
Frontier server running locally (REST on `:8000`, ConnectRPC on `:8002`):

`apps/admin/.env` (an `apps/admin/.env.example` is provided as a template):

```sh
FRONTIER_CONNECTRPC_URL=http://localhost:8002/
```

`apps/client-demo/.env`:

```sh
FRONTIER_ENDPOINT=http://localhost:8000/
FRONTIER_CONNECT_ENDPOINT=http://localhost:8002/
```

The `admin` app also reads `configs.dev.json` (served at `/configs` in dev) for
local config and terminology overrides.

## Common scripts

Run from `frontier/web`:

```sh
pnpm build      # build all packages
pnpm lint       # lint all packages
pnpm clean      # clean build outputs
pnpm changeset  # create a changeset for releases
```

The `Makefile` provides shortcuts, including an admin-only build:

```sh
make build        # install deps + build everything
make build-admin  # install deps + build only the admin app
```

## Learn more

- [Frontier documentation](https://frontier.raystack.org/)
- [`@raystack/frontier` SDK README](./sdk/README.md)
