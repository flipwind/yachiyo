<div align="center">

# Yachiyo QQ Bridge (`yachiyo-client/qq`)
_✨ Connect Project Yachiyo to QQ Groups via Official Bot API v2. ✨_

**🤖 The entire `yachiyo-client/qq/` directory was written by Antigravity (Gemini 4 Neon).**

[中文 README](README.md) · [MIT License](../../LICENSE)

</div>

## Introduction

`yachiyo-client/qq` is a standalone Go client for Project Yachiyo that bridges the **official QQ Bot API v2** to `yachiyo-gateway` (`wsjson`), enabling QQ group messaging and group-oriented tools.

It requires no changes to the main protocol, imports no internal packages other than `yachiyo-util`, and embeds a pure-Go port of `@tencent-connect/qqbot-connector` for terminal QR code binding.

## ✨ Quick Start

> [!IMPORTANT]
> Before starting, please note:
> - **Local first**: Make sure your `yachiyo-runtime` server is running on a trusted machine or network.
> - **Full group messages**: This client operates in full group message mode (`GROUP_MESSAGE_CREATE`) by default. Enable "Receive all group messages" in the QQ Open Platform and group bot settings.
> - **Non-admin defaults**: The bot is assumed not to be a group admin — muting and `@everyone` are disabled, and recalling is restricted to the bot's own messages.

### 1. Download or Build

- Download the `yachiyo-qq` archive for your platform (Linux / Windows) from the **Releases** page;
- Or build from the repository root using `Taskfile`:
  ```bash
  task qq:build
  ```
  Artifacts are written to `build/qq/`. You can also build and run in one command:
  ```bash
  task qq
  ```

### 2. Configure & Run

1. Copy `config.example.yaml` to `config.yaml` (or let the client generate a default one on first launch).
2. Adjust `yachiyo.gateway_url` if needed (default: `ws://127.0.0.1:16899/ws/`):
   - `yachiyo.client_id`: **Empty by default**. If empty on startup, a UUIDv7 is automatically generated and saved back to `config.yaml`. If a manually configured UUID is invalid, the client exits immediately before connecting.
   - `qq.app_id` / `qq.app_secret`: If empty, the client automatically launches the **embedded terminal QR bind flow** (scan with Mobile QQ to decrypt and persist credentials into `config.yaml`).
3. Run:
   ```bash
   ./yachiyo-qq -config config.yaml
   ```

### CLI Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-config` | `config.yaml` | Path to the YAML configuration file |
| `-qr` | `false` | Force Mobile QQ QR code binding to refresh `app_id` / `app_secret` |
| `-debug` | `false` | Enable debug logging |

## Protocol Mapping & Behavior

- **Session Routing**:
  - QQ groups map to `path: "group/<group_openid>"` (`wsjson://<client_id>/group/<group_openid>`).
  - QQ private (C2C) chats map to `path: "private/<user_openid>"` (`wsjson://<client_id>/private/<user_openid>`).
- **Inbound (QQ → Yachiyo)**:
  - Subscribes to `GROUP_MESSAGE_CREATE` (with `GROUP_AT_MESSAGE_CREATE` deduplicated by message ID) and `C2C_MESSAGE_CREATE` (plus `FRIEND_ADD` / `FRIEND_DEL` / `C2C_MSG_REJECT` / `C2C_MSG_RECEIVE`).
  - Formats `author` as `name [openid:...]` (falls back to `user [openid:...]` when username is empty).
  - Replaces inbound group `<@member_openid>` / `<@all>` mentions with readable `@username` / `@全体成员` (falls back to `@user [openid:...]` if uncached).
  - Converts attachments (images, audio, video, files) into readable placeholders (e.g., `[图片]`, `[文件: xxx]`).
- **Outbound (Yachiyo → QQ)**:
  - Normal replies go through Runtime's built-in `commit_message`, sent as **Markdown (`msg_type=2`)** by default; automatically falls back to plain text (`msg_type=0`) if Markdown is rejected by platform permissions.
  - Multi-bubble pacing matches `Pancake!`: **the first message is sent immediately**, and each subsequent bubble waits `clamp(300ms, 5000ms, 200 + 100 * runes ms)`.
  - Passes `<qqbot-at-user id="<member_openid>" />` tags through verbatim without outbound text replacement.
  - Manages a 60-minute passive reply `msg_id` / `msg_seq` window per group/private session (default max 5 passive replies per `msg_id`), automatically falling back to proactive messages when exhausted or for initiative messages.
- **SQLite Persistence & Size Limit**:
  - Persists groups, private sessions, member nicknames, and message history to a local SQLite database (`cache.db_path`, default `qq_cache.db`) so restarts and upgrades preserve state.
  - Enforces a configurable storage cap via `cache.max_size_mb` (default `60` MB), automatically pruning the oldest messages and reclaiming space incrementally when exceeded.

## Exposed Tools

| Tool Name | Description |
| :--- | :--- |
| `qq_list_groups` | Lists known QQ groups with `group_openid`, full sub-path `path` (e.g., `group/<group_openid>`), and full `address` |
| `qq_list_privates` | Lists known QQ private (C2C) chats with `user_openid`, full sub-path `path` (e.g., `private/<user_openid>`), and full `address` |
| `qq_get_recent_messages` | Fetches up to 100 locally cached recent messages (both user and bot replies) for a group or private chat via `path` (`group/...` / `private/...`) or `group_openid` |
| `qq_reply_group_message` | Sends a quoted reply to a specific `message_id` (prompted to avoid unless strictly necessary; prefer `commit_message`) |
| `qq_recall_group_message` | Recalls a message sent by this bot (defaults to the bot's last message in the group when `message_id` is omitted) |
| `qq_get_group_info` | Queries group metadata and bot status (falls back to local cache if restricted) |
| `qq_get_member_info` | Queries a member's details by `member_openid` (falls back to local cache if restricted) |

## License

Licensed under the [MIT License](../../LICENSE) as part of Project Yachiyo.
