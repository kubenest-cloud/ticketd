<p align="center">
    <img src="logo.png" width="120" alt="TicketD Logo"/>
</p>

<h1 align="center">TicketD</h1>

<p align="center">
    <strong>Truly lightweight. No bullshit. Self-hosted contact/ticket system in Go.</strong>
    <br />
    <br />
    <a href="buymeacoffee.com/kubenest"><img src="https://img.shields.io/badge/buy_me_a_coffee-yellow?logo=buymeacoffee&style=for-the-badge&logoColor=black&labelColor=lightblue&logoSize=auto" alt="Buy me a coffee!"></a>
</p>

<p align="center">
    <a href="https://github.com/kubenest-cloud/ticketd/releases"><img src="https://img.shields.io/github/v/release/kubenest-cloud/ticketd" alt="Release"></a>
    <a href="https://github.com/kubenest-cloud/ticketd/blob/main/LICENSE"><img src="https://img.shields.io/github/license/kubenest-cloud/ticketd" alt="License"></a>
    <a href="https://goreportcard.com/report/github.com/kubenest-cloud/ticketd"><img src="https://goreportcard.com/badge/github.com/kubenest-cloud/ticketd" alt="Go Report Card"></a>
    <a href="https://github.com/kubenest-cloud/ticketd/actions"><img src="https://github.com/kubenest-cloud/ticketd/workflows/CI/badge.svg" alt="CI Status"></a>
    <a href="https://pkg.go.dev/github.com/kubenest-cloud/ticketd"><img src="https://pkg.go.dev/badge/github.com/kubenest-cloud/ticketd.svg" alt="Go Reference"></a>
</p>

<p align="center">
    Drop an embeddable form on any site, review submissions in a tiny admin UI,<br>
    store everything locally in SQLite — no cloud, no surprises, no ongoing costs.
</p>

---

## 📋 Table of Contents

- [Features](#-features)
- [Why TicketD?](#-why-ticketd)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [Use Cases](#-use-cases)
- [Architecture](#-architecture)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)
- [Acknowledgments](#-acknowledgments)

---

## ✨ Features

- 🪶 **Ultra Lightweight**: Single binary, ~10MB, minimal memory footprint
- 🔒 **Privacy First**: Data stays on your machine, no third-party processors
- 💾 **SQLite Storage**: Single-file database, easy backups, no complex setup
- 🎨 **Embeddable Forms**: Drop a `<script>` tag anywhere, form renders instantly
- 🛡️ **CORS Protection**: Domain-based access control for each client
- 📊 **Clean Admin UI**: Modern Bulma-based dashboard for managing submissions
- 🎯 **Multiple Form Types**: Support forms (with priority) and contact forms
- 🚀 **Easy Deployment**: No dependencies beyond Go and SQLite
- 🔄 **Real-time Ready**: Structured logging with JSON output for monitoring
- ♿ **Accessible**: WCAG 2.1 AA compliant admin interface

---

## 🎯 Why TicketD?

| Aspect             | TicketD           | Cloud Services    |
| ------------------ | ----------------- | ----------------- |
| **Cost**           | $0 (only hosting) | $29-99/month      |
| **Setup**          | 2 minutes         | 30+ minutes       |
| **Data Privacy**   | Your server       | Third-party       |
| **Vendor Lock-in** | None              | High              |
| **Customization**  | Full control      | Limited           |
| **Offline**        | Works offline     | Requires internet |

### Perfect For:

- **MVPs & Betas**: Let users reach out during early product stages without cloud bills
- **Marketing Sites**: Quick contact forms for landing pages or demos
- **Internal Tools**: Lightweight ticketing for small teams or admin workflows
- **Hackathons**: Fast to deploy, minimal setup, no credit card required
- **Privacy-Conscious Projects**: Keep user data on your infrastructure

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** ([download](https://go.dev/dl/))
- **SQLite** (via cgo, requires C compiler)

### Run in 30 Seconds

```bash
# Clone the repository
git clone https://github.com/kubenest-cloud/ticketd.git
cd ticketd

# Set credentials
export TICKETD_ADMIN_USER=admin
export TICKETD_ADMIN_PASS=your-secret-password

# Run
go run .
```

Open `http://localhost:8080/admin` and start creating forms! 🎉

---

## 📦 Installation

### Option 1: From Source

```bash
git clone https://github.com/kubenest-cloud/ticketd.git
cd ticketd
go build -o ticketd
./ticketd
```

### Option 2: Using Go Install

```bash
go install github.com/kubenest-cloud/ticketd@latest
ticketd
```

### Option 3: Docker

```bash
docker build -t ticketd .
docker run -p 8080:8080 \
  -e TICKETD_ADMIN_USER=admin \
  -e TICKETD_ADMIN_PASS=secret \
  -v $(pwd)/data:/data \
  ticketd
```

### Option 4: Download Binary

Download pre-built binaries from the
[releases page](https://github.com/kubenest-cloud/ticketd/releases).

---

## ⚙️ Configuration

TicketD is configured via environment variables or a `.env` file.

### Required Variables

| Variable             | Description              | Example                |
| -------------------- | ------------------------ | ---------------------- |
| `TICKETD_ADMIN_USER` | Admin dashboard username | `admin`                |
| `TICKETD_ADMIN_PASS` | Admin dashboard password | `your-secret-password` |

### Optional Variables

| Variable                  | Default       | Description                                                 |
| ------------------------- | ------------- | ----------------------------------------------------------- |
| `TICKETD_PORT`            | `8080`        | HTTP server port                                            |
| `TICKETD_DB_PATH`         | `ticketd.db`  | SQLite database file path                                   |
| `TICKETD_PUBLIC_BASE_URL` | Auto-detected | Public URL for embed scripts (recommended in production)    |
| `TICKETD_CUSTOM_CSS`      | None          | Path to custom CSS file for embedded forms                  |
| `TICKETD_DISABLE_AUTH`    | `false`       | Disable built-in authentication (for external auth proxies) |

### Example `.env` File

```bash
TICKETD_ADMIN_USER=admin
TICKETD_ADMIN_PASS=super-secret-password
TICKETD_PORT=8080
TICKETD_DB_PATH=/var/lib/ticketd/ticketd.db
TICKETD_PUBLIC_BASE_URL=https://tickets.example.com
```

### Configuration Validation

TicketD validates configuration on startup:

- ✅ Required fields are present
- ✅ Port number is valid (1-65535)
- ✅ Custom CSS file exists (if specified)
- ✅ Database path is writable

---

## 📖 Usage

### 1. Access Admin Dashboard

Navigate to `http://localhost:8080/admin` and log in with your credentials.

### 2. Create a Client

A **client** represents a website or product. Each client has an **allowed domain** for
CORS protection.

**Example:**

- **Name**: My Awesome App
- **Allowed Domain**: `example.com` (accepts submissions from `example.com` and
  `*.example.com`)

### 3. Create a Form

After creating a client, create a **form**:

**Form Types:**

- **Support**: Includes name, email, subject, message, and priority fields
- **Contact**: Includes name, email, subject, and message fields

### 4. Embed the Form

Copy the generated embed code:

```html
<script src="https://tickets.example.com/embed/123.js"></script>
```

Paste it anywhere on your website. The form will render automatically!

#### Embedding in React/SPA Applications

For React, Next.js, Vue, or other single-page applications, use the `data-ticketd-container` attribute to specify where the form should render:

**React Example:**

```jsx
import { useEffect } from 'react';

function ContactPage() {
  useEffect(() => {
    const script = document.createElement('script');
    script.src = 'https://tickets.example.com/embed/123.js';
    script.async = true;
    document.body.appendChild(script);

    return () => {
      document.body.removeChild(script);
    };
  }, []);

  return (
    <div>
      <h1>Contact Us</h1>
      <div data-ticketd-container></div>
    </div>
  );
}
```

**Next.js Example:**

```jsx
import Script from 'next/script';

export default function ContactPage() {
  return (
    <div>
      <h1>Contact Us</h1>
      <div data-ticketd-container></div>
      <Script src="https://tickets.example.com/embed/123.js" />
    </div>
  );
}
```

**Key Points:**
- Add `data-ticketd-container` attribute to the element where you want the form
- The script will automatically find and use this container
- If no container is specified, it falls back to inserting next to the script tag

#### Troubleshooting CORS Issues

If you see a "CORS Missing Allow Origin" or "forbidden domain" error:

1. **Check the Client's Allowed Domain** in the admin dashboard:
   - For `localhost` development: Set allowed domain to `localhost`
   - For production: Use your domain without protocol (e.g., `example.com` or `mysite.com`)
   - Subdomains are automatically allowed (e.g., `example.com` allows `www.example.com`, `app.example.com`)

2. **Common Configurations**:
   ```
   Testing locally:           localhost
   Production site:           example.com
   Specific subdomain only:   app.example.com
   ```

3. **Enable Debug Logging** to see detailed CORS information:
   ```bash
   export TICKETD_DEBUG=1
   ./ticketd
   ```

4. **Localhost Port Handling**: The system automatically strips ports from localhost URLs, so `localhost` will match `localhost:3000`, `localhost:5173`, etc.

### 5. Manage Submissions

View and manage submissions in the admin dashboard:

- 📥 See all incoming tickets
- 🏷️ Update status (OPEN → IN PROGRESS → CLOSED)
- 🗑️ Delete spam or test submissions
- 📊 Filter and paginate results

---

## 💡 Use Cases

### 1. MVP Feedback Collection

```html
<!-- On your MVP landing page -->
<h2>Got feedback? Let us know!</h2>
<script src="https://tickets.yourstartup.com/embed/1.js"></script>
```

Collect user feedback without paying for expensive SaaS tools.

### 2. Multi-Client Ticketing

Perfect for agencies managing multiple client websites:

- Each client has its own forms and allowed domains
- Centralized admin dashboard
- All data in one place

### 3. Internal Team Support

Create an internal support form for your team:

- Lightweight alternative to Jira/Zendesk for small teams
- Self-hosted, no per-user pricing
- Full control over data

### 4. Event Registration

Use contact forms for event registrations:

- Hackathons
- Workshops
- Beta signups

---

## 🏗️ Architecture

### High-Level Overview

```
┌─────────────┐
│   Browser   │
│  (Website)  │
└──────┬──────┘
       │ <script src="/embed/123.js">
       ↓
┌─────────────┐      ┌──────────────┐
│  TicketD    │─────→│   SQLite     │
│  Server     │      │   Database   │
└──────┬──────┘      └──────────────┘
       │
       ↓
┌─────────────┐
│  Admin UI   │
│  Dashboard  │
└─────────────┘
```

### Project Structure

```
ticketd/
├── main.go                    # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── errors/               # Custom error types
│   ├── validator/            # Input validation
│   ├── store/                # Data models and interfaces
│   │   └── sqlite/           # SQLite implementation
│   └── web/                  # HTTP handlers and templates
│       ├── templates/        # HTML templates (Bulma CSS)
│       └── static/           # Static assets
├── Dockerfile                # Container build
└── README.md                 # This file
```

### Technologies

- **Backend**: Go 1.21+
- **Database**: SQLite 3
- **HTTP Router**: [chi](https://github.com/go-chi/chi)
- **Frontend**: Vanilla JavaScript, Bulma CSS
- **Logging**: Go's `log/slog` (structured JSON logging)

### Security Features

- ✅ **HTTP Basic Auth** for admin routes (or external auth proxy support)
- ✅ **CORS validation** per client (domain whitelist)
- ✅ **Parameterized SQL queries** (no SQL injection)
- ✅ **Input validation** on all user inputs
- ✅ **Auto-escaping** in Go templates (XSS protection)

### Authentication Modes

TicketD supports two authentication modes:

#### 1. Built-in Authentication (Default)

Uses HTTP Basic Authentication for the admin dashboard:

```bash
TICKETD_ADMIN_USER=admin
TICKETD_ADMIN_PASS=your-secret-password
```

Simple and secure for most deployments. No external dependencies required.

#### 2. External Authentication Proxy

For advanced deployments, you can disable built-in auth and use an external authentication
proxy:

```bash
TICKETD_DISABLE_AUTH=true
```

**Compatible with:**

- [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy) - OAuth/OIDC proxy
- [Authelia](https://www.authelia.com/) - Single sign-on
- [Authentik](https://goauthentik.io/) - Identity provider
- [Traefik ForwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) -
  Forward authentication
- Any reverse proxy with authentication middleware

**Example with oauth2-proxy:**

```yaml
# docker-compose.yml
services:
  oauth2-proxy:
    image: quay.io/oauth2-proxy/oauth2-proxy:latest
    command:
      - --http-address=0.0.0.0:4180
      - --upstream=http://ticketd:8080
      - --email-domain=yourcompany.com
      - --provider=google
    environment:
      OAUTH2_PROXY_CLIENT_ID: your-client-id
      OAUTH2_PROXY_CLIENT_SECRET: your-client-secret
      OAUTH2_PROXY_COOKIE_SECRET: random-secret-here
    ports:
      - "4180:4180"

  ticketd:
    image: ticketd:latest
    environment:
      TICKETD_DISABLE_AUTH: "true" # Disable built-in auth
      TICKETD_DB_PATH: /data/ticketd.db
    volumes:
      - ./data:/data
```

**⚠️ Security Warning**: Only use `TICKETD_DISABLE_AUTH=true` when deploying behind a
trusted authentication proxy. Never expose TicketD directly to the internet with
authentication disabled.

---

## 🛠️ Development

### Setup Development Environment

```bash
# Clone and install
git clone https://github.com/kubenest-cloud/ticketd.git
cd ticketd
go mod download

# Create .env file
cat > .env << EOF
TICKETD_ADMIN_USER=admin
TICKETD_ADMIN_PASS=dev123
TICKETD_PORT=8080
TICKETD_DB_PATH=ticketd_dev.db
TICKETD_PUBLIC_BASE_URL=http://localhost:8080
EOF

# Run in development mode
go run .
```

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (install golangci-lint first)
golangci-lint run

# Vet code
go vet ./...

# Check for common mistakes
staticcheck ./...
```

### Building

```bash
# Build for current platform
go build -o ticketd

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o ticketd-linux-amd64

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o ticketd-windows-amd64.exe

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ticketd-darwin-amd64

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ticketd-darwin-arm64
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Quick Contribution Guide

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help others learn and grow

### Good First Issues

Look for issues labeled
[`good first issue`](https://github.com/kubenest-cloud/ticketd/labels/good%20first%20issue) -
perfect for new contributors!

---

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for
details.

### What This Means

✅ **Commercial use** ✅ **Modification** ✅ **Distribution** ✅ **Private use**

⚠️ **No warranty or liability**

---

## 🙏 Acknowledgments

### Built With

- [Go](https://go.dev/) - The programming language
- [SQLite](https://www.sqlite.org/) - The database
- [chi](https://github.com/go-chi/chi) - HTTP router
- [Bulma](https://bulma.io/) - CSS framework
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader

### Inspired By

TicketD was built out of frustration with expensive, bloated ticketing systems for simple
use cases.

### Contributors

Thanks to all
[contributors](https://github.com/kubenest-cloud/ticketd/graphs/contributors) who have
helped make TicketD better!

---

## 📊 Project Stats

![GitHub stars](https://img.shields.io/github/stars/kubenest-cloud/ticketd?style=social)
![GitHub forks](https://img.shields.io/github/forks/kubenest-cloud/ticketd?style=social)
![GitHub watchers](https://img.shields.io/github/watchers/kubenest-cloud/ticketd?style=social)

---

## 🔗 Links

- **Documentation**: [GitHub Wiki](https://github.com/kubenest-cloud/ticketd/wiki)
  _(coming soon)_
- **Issue Tracker**: [GitHub Issues](https://github.com/kubenest-cloud/ticketd/issues)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md) _(coming soon)_
- **Roadmap**: [GitHub Projects](https://github.com/kubenest-cloud/ticketd/projects)
  _(coming soon)_

---

## 💬 Support

- 🐛 **Found a bug?**
  [Open an issue](https://github.com/kubenest-cloud/ticketd/issues/new?template=bug_report.md)
- 💡 **Have a feature request?**
  [Open an issue](https://github.com/kubenest-cloud/ticketd/issues/new?template=feature_request.md)

---

## 🌟 Show Your Support

If you find TicketD useful, please consider:

- ⭐ **Starring** the repository
- 🐦 **Sharing** on social media
- 📝 **Writing** a blog post about your experience
- 💸 **Sponsoring** the project

---

<p align="center">
    <strong>Small, focused, and honest — that's TicketD.</strong>
</p>

<p align="center">
    Made with ❤️ by the open source community
</p>
