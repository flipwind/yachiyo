<div align="center">

# Yachiyo QQ Bridge (`yachiyo-client/qq`)
_✨ Connect Project Yachiyo to QQ Groups via Official Bot API v2. ✨_

**🤖 整个 `yachiyo-client/qq/` 模块由 Antigravity（Gemini 4 Neon）独立编写完成。**

[English README](README.en.md) · [MIT License](../../LICENSE)

</div>

## 简介

`yachiyo-client/qq` 是 Project Yachiyo 的独立 Go 客户端，用于将 **QQ 官方机器人（API v2）** 接入 `yachiyo-gateway`（`wsjson`），实现 QQ 群聊消息收发与群聊相关 Tool 能力。

它不修改主协议，也不依赖除 `yachiyo-util` 外的内部模块，内置了 `@tencent-connect/qqbot-connector` 的 Go 移植版，可在终端中直接扫码绑定机器人。

## ✨ 快速开始

> [!IMPORTANT]
> 开始前请确认以下事项：
> - **本地优先**：请确保你的 `yachiyo-runtime` 服务端已在本机或受信任网络中启动。
> - **全量群消息**：本客户端默认按全量群聊消息模式工作（`GROUP_MESSAGE_CREATE`），请在 QQ 开放平台及群机器人设置中开启「接收群内全部消息」。
> - **非群管默认设定**：客户端默认按机器人非群管理员设计，不支持禁言或 `@全体成员`，撤回仅限机器人自己发出的消息。

### 1. 获取或编译程序

- 从 **Releases** 页面下载对应平台（Linux / Windows）的 `yachiyo-qq` 压缩包；
- 或在项目根目录直接运行 `Taskfile` 命令编译：
  ```bash
  task qq:build
  ```
  构建产物会输出到 `build/qq/` 目录。也可以直接一键构建并运行：
  ```bash
  task qq
  ```

### 2. 配置与启动

1. 复制 `config.example.yaml` 为 `config.yaml`（若未创建，首次启动时也会自动生成默认配置）。
2. 按需填写 `yachiyo.gateway_url`（默认 `ws://127.0.0.1:16899/ws/`）：
   - `yachiyo.client_id`：**默认留空**。首次启动为空时会自动生成 UUIDv7 并回写保存；若手动填写了非法 UUID，程序会在连接前直接报错退出。
   - `qq.app_id` / `qq.app_secret`：若留空，首次启动时会自动拉起**终端扫码绑定**（手机 QQ 扫码确认后自动解密凭据并回写到 `config.yaml`）。
3. 运行：
   ```bash
   ./yachiyo-qq -config config.yaml
   ```

### 命令行参数

| 参数 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `-config` | `config.yaml` | 指定 YAML 配置文件路径 |
| `-qr` | `false` | 强制重新执行手机 QQ 扫码绑定并刷新 `app_id` / `app_secret` |
| `-debug` | `false` | 开启 Debug 级别日志 |

## 协议映射与行为说明

- **会话路由**：
  - QQ 群聊通过 `path: "group/<group_openid>"` 路由，完整地址为 `wsjson://<client_id>/group/<group_openid>`。
  - QQ 私聊通过 `path: "private/<user_openid>"` 路由，完整地址为 `wsjson://<client_id>/private/<user_openid>`。
- **入站消息（QQ → Yachiyo）**：
  - 群聊订阅 `GROUP_MESSAGE_CREATE`（同时兜底 `GROUP_AT_MESSAGE_CREATE` 并按消息 ID 去重）；私聊订阅 `C2C_MESSAGE_CREATE`（以及 `FRIEND_ADD` / `FRIEND_DEL` / `C2C_MSG_REJECT` / `C2C_MSG_RECEIVE`）。
  - `author` 统一格式化为 `name [openid:...]`（昵称为空时回退 `user [openid:...]`）。
  - 群聊入站文本自动将 `<@member_openid>` / `<@all>` 替换为可读的 `@username` / `@全体成员`（未缓存昵称时回退 `@user [openid:...]`）。
  - 图片、语音、视频、文件等附件转为可读占位符（如 `[图片]`、`[文件: xxx]`）。
- **出站消息（Yachiyo → QQ）**：
  - 常规回复通过 Runtime 内置的 `commit_message` 发送，默认走 **Markdown（`msg_type=2`）**；若遇平台无 Markdown 权限错误，自动降级为纯文本（`msg_type=0`）重发。
  - 多条消息气泡与 `Pancake!` 保持一致的节奏：**首条消息立即发送**，后续消息按字数间隔 `clamp(300ms, 5000ms, 200 + 100 * 字数 ms)` 依次发送。
  - 出站正文原样透传官方 `<qqbot-at-user id="<member_openid>" />` 标签，不做文本替换。
  - 自动维护每群/每私聊会话 60 分钟被动回复 `msg_id` / `msg_seq` 窗口（默认单条 `msg_id` 最多被动回复 5 次），超出额度或主动消息时自动回退为主动发送。
- **SQLite 持久化与空间上限**：
  - 群列表、私聊会话、群成员昵称与消息历史均持久化至本地 SQLite（`cache.db_path`，默认 `qq_cache.db`），重启或升级不丢失状态。
  - 通过 `cache.max_size_mb`（默认 `60` MB）控制全库存储上限，超限时自动滚动淘汰最旧的历史消息并增量回收空间。

## 暴露给 Yachiyo 的 Tools

| Tool 名称 | 说明 |
| :--- | :--- |
| `qq_list_groups` | 列出已知 QQ 群，返回 `group_openid`、完整子路径 `path`（如 `group/<group_openid>`）与完整 `address` |
| `qq_list_privates` | 列出已知 QQ 私聊会话，返回 `user_openid`、完整子路径 `path`（如 `private/<user_openid>`）与完整 `address` |
| `qq_get_recent_messages` | 获取某群或某私聊本地缓存的最近消息（含用户与机器人回复，最多 100 条），支持 `path`（`group/...` / `private/...`）或 `group_openid` |
| `qq_reply_group_message` | 带引用气泡回复某条群消息（Prompt 引导尽量不用，常规发言优先 `commit_message`） |
| `qq_recall_group_message` | 撤回机器人自己在群内发出的消息（省略 `message_id` 时默认撤回本群上一条） |
| `qq_get_group_info` | 查询群基础信息及机器人群内状态（接口受限时降级返回本地缓存） |
| `qq_get_member_info` | 按 `member_openid` 查询群成员信息（接口受限时降级返回本地缓存） |

## 协议

本模块随 Project Yachiyo 一并以 [MIT License](../../LICENSE) 开源。
