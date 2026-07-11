# CLIProxyAPI Fork Notes

This fork follows upstream CLIProxyAPI. Most baseline API, model, OAuth, translator, and runtime behavior comes from upstream syncs and is intentionally not repeated here.

This README only documents the behavior this fork currently changes on top of the synced codebase.

Upstream project: [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)

Upstream documentation: [help.router-for.me](https://help.router-for.me/)

## What This Fork Changes

Compared with the synced upstream code, this fork adds Claude cloaking controls and changes repository operations:

- Adds configurable Claude OAuth tool-name renaming through `claude-tool-rename`.
- Adds `cloak.preserve-system-prompt` so cloaked Claude requests can keep the caller's original system prompt when desired.

## Claude Tool-Name Rename

Claude OAuth cloaking renames tool names before sending requests upstream, then maps response tool names back to the client-visible names. This helps non-official client tools look closer to official Claude Code traffic.

The fork keeps the built-in rename map and adds a top-level `claude-tool-rename` map for extra tools or overrides.

```yaml
claude-tool-rename:
  call_omo_agent: callOmoAgent
  my_custom_tool: myCustomTool
```

Usage notes:

- Keys are incoming client tool names.
- Values are the tool names sent upstream.
- Custom entries extend the built-in map.
- If a key already exists in the built-in map, the configured value overrides it.
- Response tool names are mapped back automatically for the request.
- Config hot reload updates the rename map without restarting the server.

Example: a client sends a Claude OAuth request with a tool named `call_omo_agent`; the proxy sends it upstream as `callOmoAgent`, then restores response references back to `call_omo_agent` for the client.

## Preserve System Prompt During Claude Cloaking

Claude cloaking normally rewrites forwarded system instructions so requests look more like native Claude Code traffic. This fork adds an opt-in flag to preserve the caller's original system prompt when cloaking is active and strict mode is off.

For `claude-api-key` entries, set `cloak.preserve-system-prompt`:

```yaml
claude-api-key:
  - api-key: "sk-ant-..."
    cloak:
      mode: "auto"
      strict-mode: false
      preserve-system-prompt: true
```

For Claude OAuth/file-backed auth records, set the auth attribute in the token JSON:

```json
{
  "type": "claude",
  "email": "user@example.com",
  "cloak_mode": "auto",
  "cloak_strict_mode": "false",
  "cloak_preserve_system_prompt": "true"
}
```

Behavior:

- `false` or unset keeps the default cloaking behavior: the forwarded system prompt is reduced to a neutral reminder.
- `true` keeps the original system prompt and moves it into the first user message instead of replacing it.
- The option only matters when cloaking is active and `strict-mode` is `false`.
- In `strict-mode: true`, user system messages are still stripped and only the Claude Code prompt is kept.

Tradeoff: preserving the original system prompt can make requests more recognizable as third-party traffic, which may affect Anthropic plan-vs-extra-usage metering. Keep it disabled unless you need exact system-prompt preservation.

## Docker Image

The fork image is published under:

```text
ghcr.io/pedxyuyuko/cli-proxy-api
```

Pull the latest image:

```bash
docker pull ghcr.io/pedxyuyuko/cli-proxy-api:latest
```

Run it with a local config file:

```bash
docker run --rm \
  -p 8317:8317 \
  -v "$PWD/config.yaml:/app/config.yaml" \
  ghcr.io/pedxyuyuko/cli-proxy-api:latest
```

Use a tagged release image:

```bash
docker pull ghcr.io/pedxyuyuko/cli-proxy-api:<tag>
docker run --rm -p 8317:8317 ghcr.io/pedxyuyuko/cli-proxy-api:<tag>
```

Replace `<tag>` with the Git tag that triggered the workflow.

## Local Development

Use the normal upstream commands for local development.

```bash
cp config.example.yaml config.yaml
go run ./cmd/server --config config.yaml
```

For full configuration and feature usage, use the upstream docs instead of this fork README:

- [CLIProxyAPI guide](https://help.router-for.me/)
- [Management API](https://help.router-for.me/management/api)

## Notes For Contributors To This Fork

- Keep this README limited to fork-only differences.

