<p align="center">
    <img src="logo.png" width="120"/>
</p>

# TicketD

TRULY lightweight. No bullshit. Self-hosted contact/ticket system in Go.

Drop an embeddable form on any site, review submissions in a tiny admin UI,
store everything locally in SQLite — no cloud, no surprises, no ongoing costs.

## Why TicketD?

- **Minimal:** Small binary, single SQLite file, negligible compute needs.
- **Private:** Data stays on your machine — no third-party processors.
- **Cheap:** Perfect for MVPs, betas, demos and hackathons — no cloud bills.
- **Simple:** Easy to embed, easy to manage, no ops overhead.

## Use cases

- **MVP / Beta feedback:** Let users reach out during early product stages
  without spending cloud credits.
- **Marketing/demo sites:** Quick contact forms for landing pages or demos.
- **Internal tools / ops:** Lightweight ticketing for small teams or admin-only
  workflows.
- **Hackathons & prototypes:** Fast to deploy, minimal setup.

## Requirements

- Go 1.21+
- SQLite (via `mattn/go-sqlite3`)
- A C compiler (required by `go-sqlite3` because it uses cgo; on Windows install
  MSYS2 or TDM-GCC)

## Configure

Set these environment variables before running:

- `TICKETD_ADMIN_USER` and `TICKETD_ADMIN_PASS` (required)
- `TICKETD_PORT` (default: `8080`)
- `TICKETD_DB_PATH` (default: `ticketd.db`)
- `TICKETD_PUBLIC_BASE_URL` (recommended in production, e.g.
  `https://ticketd.example.com`)
- `TICKETD_CUSTOM_CSS` (optional path to a custom CSS file for embedded forms)

TicketD will also load a local `.env` file if present.

## Run

```bash
TICKETD_ADMIN_USER=admin \
TICKETD_ADMIN_PASS=secret \
TICKETD_PUBLIC_BASE_URL=http://localhost:8080 \
go run .
```

If you see `go-sqlite3 requires cgo to work`, enable cgo and ensure a C compiler
is installed, on Windows for example:

```ps1
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
$env:CC  = "C:\msys64\ucrt64\bin\gcc.exe"
$env:CXX = "C:\msys64\ucrt64\bin\g++.exe"
$env:CGO_ENABLED=1
go run .
```

Open the admin dashboard at `http://localhost:8080/admin`.

## Embed a form

After creating a client and a form in the admin UI, copy the generated script
URL:

```html
<script src="https://ticketd.example.com/embed/123.js"></script>
```

The form will render right where the script tag is placed.

## Notes

- Each client has an `allowed domain`. Submissions must originate from that
  domain.
- The default form style is bundled, but you can replace it via
  `TICKETD_CUSTOM_CSS`.


Small, focused, and honest — that's TicketD.
<p align="center">
    <img src="logo.png" width="120"/>
</p>

# TicketD

TicketD is a lightweight, self-hosted ticketing/contact system built in Go. It
provides embeddable forms for your public sites and a simple admin dashboard to
review submissions.

## Requirements

- Go 1.21+
- SQLite (via `mattn/go-sqlite3`)
- A C compiler (required by `go-sqlite3` because it uses cgo; on Windows install
  MSYS2 or TDM-GCC)

## Configure

Set these environment variables before running:

- `TICKETD_ADMIN_USER` and `TICKETD_ADMIN_PASS` (required)
- `TICKETD_PORT` (default: `8080`)
- `TICKETD_DB_PATH` (default: `ticketd.db`)
- `TICKETD_PUBLIC_BASE_URL` (recommended in production, e.g.
  `https://ticketd.example.com`)
- `TICKETD_CUSTOM_CSS` (optional path to a custom CSS file for embedded forms)

TicketD will also load a local `.env` file if present.

## Run

```bash
TICKETD_ADMIN_USER=admin \
TICKETD_ADMIN_PASS=secret \
TICKETD_PUBLIC_BASE_URL=http://localhost:8080 \
go run .
```

If you see `go-sqlite3 requires cgo to work`, enable cgo and ensure a C compiler
is installed, on windows, for example:

```ps1
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
$env:CC  = "C:\msys64\ucrt64\bin\gcc.exe"
$env:CXX = "C:\msys64\ucrt64\bin\g++.exe"
$env:CGO_ENABLED=1
go run .
```

Open the admin dashboard at `http://localhost:8080/admin`.

## Embed a form

After creating a client and a form in the admin UI, copy the generated script
URL:

```html
<script src="https://ticketd.example.com/embed/123.js"></script>
```

The form will render right where the script tag is placed.

## Notes

- Each client has an `allowed domain`. Submissions must originate from that
  domain.
- The default form style is bundled, but you can replace it via
  `TICKETD_CUSTOM_CSS`.
