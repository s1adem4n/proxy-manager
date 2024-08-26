# Proxy Manager

*Minimalistic reverse proxy manager for Caddy*

## Usage
Ensure that Caddy is running and the API is enabled. Try going to `http://localhost:2019/` to check if the API is enabled.
For Podman support, please enable/start the Podman socket.

**Flags**
- `--address` - The address to bind the proxy manager to. Default: `:8080`
- `--podman` - Whether to scan for container labels using the Podman socket. Possible labels are `proxy-manager.enable` and `proxy-manager.name`. Default: `false`
- `--domain` - Default domain used when creating proxies for containers. Default: `example.com`

## Controlling a remote Caddy instance
Changing the Caddy API url is not possible, as you should *never* expose your Caddy API to the internet. Instead, you can use a reverse SSH tunnel to access the API. Here's an example:

```bash
ssh -R 2019:localhost:2019 user@remote
```