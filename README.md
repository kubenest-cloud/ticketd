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
