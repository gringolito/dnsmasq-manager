# dnsmasq-manager

[![Build](https://github.com/gringolito/dnsmasq-manager/actions/workflows/tests.yaml/badge.svg)](https://github.com/gringolito/dnsmasq-manager/actions/workflows/tests.yaml)
[![Coverage](https://codecov.io/gh/gringolito/dnsmasq-manager/graph/badge.svg)](https://codecov.io/gh/gringolito/dnsmasq-manager)
[![Go Version](https://img.shields.io/github/go-mod/go-version/gringolito/dnsmasq-manager)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/gringolito/dnsmasq-manager)](https://github.com/gringolito/dnsmasq-manager/releases/latest)
[![License](https://img.shields.io/badge/license-Beer--ware-yellow)](LICENSE)

A lightweight REST API for managing static DHCP/DNS entries on a [dnsmasq](https://thekelleys.org.uk/dnsmasq/doc.html) server.

Instead of hand-editing dnsmasq configuration files, `dnsmasq-manager` exposes a simple HTTP API so that other tools, scripts, or dashboards can manage static leases programmatically — no SSH required.

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- Manage static DHCP host reservations — add, list, update, and delete
- Query hosts by MAC address or IP address
- JWT authentication with multiple algorithm support (ECDSA, RSA, HMAC)
- Role-scoped authorization (`dhcp:read`, `dhcp:add`, `dhcp:change`, `dhcp:admin`)
- Interactive OpenAPI / Swagger UI included out of the box
- Structured JSON or plain-text logging with configurable severity level
- Systemd service unit with a hardened security profile
- Multi-architecture Linux packages: `.deb`, `.rpm`, `.apk`, ArchLinux

---

## Requirements

- Linux system with systemd
- [dnsmasq](https://thekelleys.org.uk/dnsmasq/doc.html) installed and managing `/etc/dnsmasq.d/`
- Go 1.24+ (source builds only)

---

## Installation

### From a package (recommended)

Download the package for your distribution from the [latest release](https://github.com/gringolito/dnsmasq-manager/releases/latest).

**Debian / Ubuntu**
```bash
wget https://github.com/gringolito/dnsmasq-manager/releases/latest/download/dnsmasq-manager_<version>_amd64.deb
sudo dpkg -i dnsmasq-manager_<version>_amd64.deb
```

**RHEL / Fedora**
```bash
sudo rpm -i dnsmasq-manager_<version>_amd64.rpm
```

**Alpine**
```bash
sudo apk add --allow-untrusted dnsmasq-manager_<version>_amd64.apk
```

Packages for `arm`, `arm64`, `armv5/6/7`, and `i386` are also available on the releases page.

### From a tarball

```bash
wget https://github.com/gringolito/dnsmasq-manager/releases/latest/download/dnsmasq-manager_Linux_x86_64.tar.gz
tar -xzf dnsmasq-manager_Linux_x86_64.tar.gz
sudo mv dnsmasq-manager /usr/local/bin/
```

### From source

```bash
git clone https://github.com/gringolito/dnsmasq-manager
cd dnsmasq-manager
go build -tags=release -o dnsmasq-manager .
```

---

## Configuration

The configuration file is located at `/etc/dnsmasq-manager/config.yaml` (installed automatically by the package). A fully annotated template is provided at `config.yaml.dist`.

All settings have sensible defaults, so the service works out of the box with no changes required. Uncomment and adjust only what you need:

```yaml
# /etc/dnsmasq-manager/config.yaml

# Path to the dnsmasq static DHCP leases file.
# Default: /etc/dnsmasq.d/04-dhcp-static-leases.conf
#
# host:
#   static:
#     file: /etc/dnsmasq.d/04-dhcp-static-leases.conf

# JWT-based authentication for API endpoints.
# Available methods: none, ecdsa-256, ecdsa-384, ecdsa-512,
#                    hmac-256, hmac-384, hmac-512,
#                    rsa-256, rsa-384, rsa-512
# The key can be a file path (ECDSA/RSA public key) or a plain-text secret (HMAC).
# Default: no authentication
#
# auth:
#   method: ecdsa-512
#   key: /etc/dnsmasq-manager/id_ecdsa.pub

# HTTP listening port.
# Default: 6904
#
# server:
#   port: 6904

# Logging settings.
# Available levels: debug, info, warning, error
# Available formats: json, text
# Default: info / json / stdout
#
# log:
#   level: info
#   format: json
#   file: /var/log/dnsmasq-manager.log
#   source: false
```

### Environment variables

Every configuration key can be overridden with an environment variable using the `DMM_` prefix and underscores for nesting:

| Variable | Default | Description |
|---|---|---|
| `DMM_SERVER_PORT` | `6904` | HTTP listening port |
| `DMM_AUTH_METHOD` | `none` | JWT algorithm |
| `DMM_AUTH_KEY` | — | JWT key path or secret |
| `DMM_LOG_LEVEL` | `info` | Log verbosity |
| `DMM_LOG_FORMAT` | `json` | Log output format |
| `DMM_LOG_FILE` | — | Log file path (stdout if empty) |

### Setting up JWT authentication

The package ships with a key generation helper:

```bash
# Generate an ECDSA-512 key pair in the current directory
generate-jwt-keys ecdsa-512
```

Point the `auth.key` config option at the generated public key file, then issue JWTs signed with the corresponding private key. Include the appropriate scope claim (`dhcp:read`, `dhcp:add`, `dhcp:change`, or `dhcp:admin`) to control what each token can do.

---

## Usage

### Starting the service

```bash
sudo systemctl enable --now dnsmasq-manager
```

Check the service status and logs:

```bash
sudo systemctl status dnsmasq-manager
sudo journalctl -u dnsmasq-manager -f
```

### API examples

**List all static hosts**
```bash
curl http://localhost:6904/api/v1/static/hosts
```

**With JWT authentication**
```bash
TOKEN="<your-jwt>"
curl -H "Authorization: Bearer $TOKEN" http://localhost:6904/api/v1/static/hosts
```

**Get a specific host by MAC address**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:6904/api/v1/static/host?mac=aa:bb:cc:dd:ee:ff"
```

**Add a static host**
```bash
curl -X POST http://localhost:6904/api/v1/static/host \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"MacAddress":"aa:bb:cc:dd:ee:ff","IPAddress":"192.168.1.100","HostName":"mydevice"}'
```

**Update a static host**
```bash
curl -X PUT http://localhost:6904/api/v1/static/host \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"MacAddress":"aa:bb:cc:dd:ee:ff","IPAddress":"192.168.1.101","HostName":"mydevice"}'
```

**Delete a host**
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  "http://localhost:6904/api/v1/static/host?mac=aa:bb:cc:dd:ee:ff"
```

### Swagger UI

Full interactive API documentation is available at:

```
http://<host>:6904/openapi/swagger-ui
```

---

## API Reference

| Method | Path | Required scope | Description |
|---|---|---|---|
| `GET` | `/api/v1/static/hosts` | `dhcp:read` | List all static hosts |
| `GET` | `/api/v1/static/host?mac=` \| `?ip=` | `dhcp:read` | Get a host by MAC or IP |
| `POST` | `/api/v1/static/host` | `dhcp:add` | Add a new static host |
| `PUT` | `/api/v1/static/host` | `dhcp:change` | Update an existing host |
| `DELETE` | `/api/v1/static/host?mac=` \| `?ip=` | `dhcp:change` | Remove a host |
| `GET` | `/metrics` | — | Server metrics |

The raw OpenAPI spec is served at `/openapi/spec`.

---

## Roadmap

### MVP (complete)

- [x] Static DHCP host management (add, list, update, delete)
- [x] JWT authentication and role-based authorization
- [x] Configurable via YAML and environment variables
- [x] Structured logging
- [x] Systemd service with hardened security profile
- [x] OpenAPI / Swagger documentation
- [x] Unit tests
- [x] Multi-architecture Linux packages (`.deb`, `.rpm`, `.apk`, ArchLinux)

### Phase 2 (planned)

- [ ] Static DNS entry management
- [ ] CNAME alias management

---

## Contributing

Contributions are welcome and appreciated. Here are a few ways to get involved:

- **Found a bug?** [Open an issue](https://github.com/gringolito/dnsmasq-manager/issues/new) with as much detail as possible — steps to reproduce, expected vs. actual behavior, and your environment.
- **Have a feature idea?** [Start a discussion via an issue](https://github.com/gringolito/dnsmasq-manager/issues/new) before writing code, so we can align on the approach.
- **Want to submit a fix or improvement?** Fork the repo, make your changes, and open a pull request. Please include tests for any new behavior and keep the scope focused.

---

## License

[The Beer-Ware License](LICENSE) — do whatever you want with this. If we ever meet and you think it was worth it, buy me a beer. 🍺
