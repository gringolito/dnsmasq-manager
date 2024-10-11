# dnsmasq-manager

Dnsmasq DNS / DHCP management API

This project provides a RESTful API to manage some of the DHCP/DNS resources on a dnsmasq
server, like:

- Manage static DHCP entries
- Manage static DNS entries
- Manage CNAME aliases

## TO-DO

### MVP

- [x] Add authentication / authorization control over the routes
- [x] Improve logging with severity classified messages
- [x] Move routes to a /api/v1 prefix
- [x] Make the system configurable (maybe use Viper?)
- [x] Make the code release-ready (~~set gin properly~~ and remove pretty JSON methods)
- [x] Create the systemd service files
- [ ] Setup a CI pipeline
- [x] Create .deb deployable package (target at least to armv6)
    - [x] .rpm, .apk, ArchLinux packages and tarball archives for all the main architectures (BONUS)
- [x] Create a OpenAI/Swagger documentation
- [x] Add unit tests

### Phase-2

- [ ] Manage static DNS entries
- [ ] Manage CNAME alias
