# SpendSense Backend (Go)

Income-first expense tracker. Set your total income, allocate fixed expenses, savings, and needs — then add variable expenses gradually to see how much you have left to spend.

Built with **Go + ConnectRPC**. Works natively with a React frontend via `@connectrpc/connect-web` — no Envoy proxy required.

> **API contract (proto files)** are maintained in the separate [`SpendSense-proto`](https://github.com/XpendSense/SpendSense-proto) repository and published to the [Buf Schema Registry](https://buf.build/xpendsense/spendsense).

---

## Stack

| Concern | Library |
|---|---|
| RPC | [ConnectRPC](https://connectrpc.com) |
| API contract | [buf.build/xpendsense/spendsense](https://buf.build/xpendsense/spendsense) |
| Database | [Neon](https://neon.tech) (serverless PostgreSQL) via `jackc/pgx/v5` |
| Queries | `sqlc` (type-safe Go from SQL) |
| Auth | JWT (`golang-jwt/jwt/v5`) + Google OAuth2 |
| Secrets | [SOPS](https://github.com/getsops/sops) + [age](https://github.com/FiloSottile/age) |
| Migrations | [goose v3](https://github.com/pressly/goose) (library mode via `cmd/migrate`) |
| CI | GitHub Actions — build, test, migrate on every push |

---

## Prerequisites

- Go 1.23+
- [buf CLI](https://buf.build/docs/installation)
- [sqlc CLI](https://docs.sqlc.dev/en/latest/overview/install.html)
- [SOPS CLI](https://github.com/getsops/sops/releases) (for secrets)
- [age CLI](https://github.com/FiloSottile/age/releases) (for generating/using encryption keys)

```powershell
go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

---

## One-time setup

### 1. Get your age private key

Secret files (`.env.dev`, `.env.prod`, etc.) are encrypted with SOPS + age. To decrypt them you need the private key. Place it at the SOPS default location:

- **Windows:** `%APPDATA%\sops\age\keys.txt`
- **Linux/macOS:** `~/.config/sops/age/keys.txt`

Ask a team member for the key, or if setting up a new environment, generate one:

```powershell
age-keygen -o "$env:APPDATA\sops\age\keys.txt"
# then share the public key so .sops.yaml can be updated
```

### 2. Decrypt secrets

```powershell
make secrets-decrypt ENV=dev
# produces .env.dev (gitignored)
```

`.env.dev` contains:

```
DATABASE_URL=postgresql://<user>:<password>@<host>.neon.tech/<db>?sslmode=require
JWT_SECRET=<a-long-random-string>
GOOGLE_CLIENT_ID=<from Google Cloud Console>
GOOGLE_CLIENT_SECRET=<from Google Cloud Console>
```

Get the Neon connection string from the [Neon console](https://console.neon.tech) — use the **direct** connection (no `-pooler` suffix).

### 3. Apply the database schema

Run the schema once against your Neon database. You can do this via the Neon SQL editor or any PostgreSQL client:

```powershell
psql $DATABASE_URL -f internal/db/migrations/schema.sql
```

The schema file is at [internal/db/migrations/schema.sql](internal/db/migrations/schema.sql).

### 4. Authenticate with the Buf Schema Registry

```powershell
buf registry login buf.build
# Enter your BSR token when prompted (generate one at buf.build/settings/user)
```

Required to pull the proto module from BSR. Without it, `buf generate` will fail even for public modules.

### 5. Generate code

```powershell
make generate
# runs: buf generate + sqlc generate
```

> `buf generate` must run before `go build` — handlers import from `gen/` which only exists after generation.

### 6. Start the server

```powershell
make run
# Server starts on http://localhost:8080
```

---

## Daily development

```powershell
make run              # start server (dev env)
make test             # go test ./...
make build            # compile binary to bin/
make generate         # buf generate + sqlc generate
make tidy             # go mod tidy

make secrets-encrypt ENV=dev   # encrypt .env.dev → .env.dev.enc (commit this)
make secrets-decrypt ENV=dev   # decrypt .env.dev.enc → .env.dev (gitignored)
```

---

## Secrets workflow

Plaintext `.env.*` files are **gitignored**. Only the encrypted `.env.*.enc` files are committed.

To update a secret:

```powershell
# 1. Edit .env.dev as needed
# 2. Re-encrypt
make secrets-encrypt ENV=dev
# 3. Commit the .enc file
git add .env.dev.enc
```

CI decrypts using the `AGE_SECRET_KEY` repository secret (key contents, not a file path).

---

## Picking up proto changes

When `SpendSense-proto` publishes a new version:

```powershell
make generate   # fetches latest proto from BSR and regenerates Go code
```

---

## CI pipeline

Every push to `main`, `develop`, or `feature/**` runs the following in order:

1. **Generate proto code** — pulls the latest proto from BSR and generates Go types
2. **Decrypt secrets** — decrypts `.env.dev.enc` using the `AGE_SECRET_KEY` repository secret
3. **Run migrations** — applies any pending migrations against the Neon dev database via `go run ./cmd/migrate up`
4. **Build** — `go build ./...`
5. **Test** — `go test ./...`

### Why `cmd/migrate` instead of a CLI tool

We use goose as a **library** (`pressly/goose/v3`) wired to a `pgx/v5` connection rather than the goose or golang-migrate CLI. The CLI tools use `lib/pq` which does not handle Neon's `channel_binding=require` connection parameter, silently falling back to a local socket. `pgx/v5` handles all Neon parameters correctly.

### Adding a GitHub secret

The CI decrypt step requires `AGE_SECRET_KEY` in GitHub repo secrets (Settings → Secrets → Actions). The value is the **full contents** of your age private key file (`%APPDATA%\sops\age\keys.txt`), not the file path.

---

## Schema changes

The database schema lives in [internal/db/migrations/schema.sql](internal/db/migrations/schema.sql). When you change it:

1. Edit `schema.sql`
2. Run `sqlc generate` to regenerate typed query methods
3. Apply the updated schema to the Neon database
4. Commit `schema.sql` and the regenerated `internal/sqlc/` files together

---

## Project structure

```
.
├── buf.yaml                    # Declares proto dependency (buf.build/xpendsense/spendsense)
├── buf.gen.yaml                # Code generation config — outputs Go types to gen/
├── buf.lock                    # Pinned proto version (commit this)
├── sqlc.yaml                   # Points at local schema + query SQL
├── .sops.yaml                  # age public key for SOPS encryption rules
│
├── gen/                        # Generated by `buf generate` — do not edit
│
├── internal/
│   ├── config/                 # Env-based configuration
│   ├── db/
│   │   ├── conn.go             # pgxpool setup (Neon-compatible settings)
│   │   ├── migrations/         # Database schema (schema.sql)
│   │   └── query/              # SQL queries consumed by sqlc
│   ├── sqlc/                   # Generated by `sqlc generate` — do not edit
│   ├── auth/                   # JWT token service + Google OAuth client
│   ├── apperr/                 # Typed error values (NotFound, Forbidden, Duplicate, Invalid)
│   ├── middleware/             # ConnectRPC interceptors (auth, logging)
│   ├── repository/             # Database access layer
│   ├── service/                # Business logic layer (unit-tested)
│   └── handler/                # RPC handler implementations
│
└── cmd/
    └── server/main.go          # Application entry point
```

---

## Connecting a React frontend

The frontend consumes pre-generated npm packages published automatically by the BSR — no local proto tooling required.

```typescript
import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { BudgetService } from "@buf/xpendsense_spendsense.connectrpc_es/spendsense/v1/budget_pb";

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem("access_token");
      if (token) req.header.set("Authorization", `Bearer ${token}`);
      return next(req);
    },
  ],
});

export const budgetClient = createClient(BudgetService, transport);
```

---

## Docker

```powershell
docker build -t spendsense-api .
docker run -p 8080:8080 --env-file .env.dev spendsense-api
```
