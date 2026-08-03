# 归档说明（Archive Notice）
**本仓库已归档，不再维护。**
这个 fork 的存在理由只有两条 Claude cloaking 定制：
- `claude-tool-rename`（commit `aea16801`）— 在内置的 lowercase→TitleCase 工具名映射表上追加/覆盖条目，用于削弱 Anthropic 对第三方客户端的 tool-name 指纹识别。
- `cloak.preserve-system-prompt`（commit `f33177d1`）— 在 cloaking 生效且 `strict-mode: false` 时，可选择保留调用方原始 system prompt，而不是替换成中性提示。
上游 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 已分别用更完整的机制取代了这两者：
- 上游 `f3e25ab2` 引入 per-request MCP tool aliasing（`helps/claude_mcp_alias.go`）。cloaked OAuth 请求上的**每一个** custom tool 都自动映射为 caller 维度稳定的不透明 `mcp__<server>__<tool>` 别名，请求内符号表在响应和流式事件中还原原名。覆盖面严格大于一张手工维护的映射表，`claude-tool-rename` 所依赖的静态 map 已被删除。
- 上游 `ef89c6a6` + `3fac4a09` 重建了 cloaked system prompt 的处理方式。`strict-mode: false` 下调用方 system 块**原样保留**，仅按模型改变位置：legacy 白名单模型走 `<system-reminder>` 前置到首个 user 消息，其余及未来模型走 mid-conversation `role=system` —— 后者本就是原生 Claude Code 的线格式。这等价于 fork 的 `preserve-system-prompt: true`，且已是默认行为；prompt 脱敏逻辑被上游判定为不再必要而移除。
结论：fork 的两个特性一个被超集覆盖，一个被扶正为上游默认。剩余 delta 仅为 Docker 镜像发布路径（`ghcr.io/pedxyuyuko/cli-proxy-api`）与 `fork/main` 分支 CI，不足以支撑一个独立 fork。
同步成本也已不成比例：最后一次上游同步（PR #17）落后 113 个提交、163 个文件、19,321 行插入，冲突集中在 fork 特性所在的 `claude_executor*.go`，且冲突双方是两套互斥的架构而非可合并的增量。
**迁移方式**：直接使用上游。删除 config 中的 `claude-tool-rename` 块与 `cloak.preserve-system-prompt` 键，以及 Claude OAuth token JSON 中的 `cloak_preserve_system_prompt` 属性——三者在上游均无对应配置项，也不需要。若原先依赖 `preserve-system-prompt: false` 的脱敏行为，上游最接近的是 `cloak.strict-mode: true`，但语义是完全丢弃调用方 prompt 而非降级，会改变模型行为。

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

