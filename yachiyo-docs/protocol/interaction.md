# Interaction

当前阶段通常指 runtime 和 client 间的消息。

其中 `category` 标识为 `interaction`。

## Client Message

指 client 向 runtime 发送的消息。  
当前**仅支持文字**。

```
type: string = "client_message"
data {
    message: string
    time: int64
    author?: string
    path?: string
}
```

`message` 内是纯文本内容。
`time` 为 unix 时间戳。
`author`（可选）为发言人昵称或标识；省略或为空时默认为 `"user"`。
`path`（可选）为自定义子路径，用于在同一 client 连接下区分不同会话（如 `group/1126376305`、`private/123456`）。例如，若该字段省略或为空时，完整地址为 `wsjson://<client_id>`；填写后完整地址为 `wsjson://<client_id>/<path>`。

> [!NOTE]
> `path` 规范：
> - **仅允许** `a-z`、`A-Z`、`0-9`、`.`、`_`、`-`、`/`，不超过 128 个字符；
> - 禁止空段（`//`）及 `.`、`..` 路径段；
> - 传入非法 `path` 时，gateway 会拒绝该消息并返回 `code: "invalid_path"` 的 `runtime_error`。

## Runtime Message

指 runtime 向 client 发送的消息。  
同样在当前**仅支持文字**。

```
type: string = "runtime_message"
data {
    message: List<string>
    is_initiative: bool
    time: int64
    path?: string
}
```

`message` 内为消息的列表，用于表示间隔地发送了数条消息。  
`is_initiative` 用于标示是否为 runtime 主动发出的消息。
`time` 为 unix 时间戳。
`path`（可选）为自定义子路径。当消息发往带 path 的地址时由 gateway 原样带回；否则省略。

## Get Relative Message History

使用该协议，runtime 会返回一个与该 client_id 相匹配的 message history 列表。

### 发送请求

```
type: string = "get_relative_message_history"
data {}
```

### 返回值

```
type: string = "relative_message_history"
data {
    messages: List<runtime_message | client_message> = [
        {
            type: string = "client_message"
            data: {
                message: string
                time: int64
            }
        }
        ...
    ]
}
```

## Error Message

由于一些这种那种的原因，我们的 runtime 最终没有处理好你的消息。

由 runtime 推送。

```
type: string = "runtime_error"
data {
    code: string
    message: string
    path?: string
}
```

`code` 是发生错误的代号。
`message` 是具体错误的内容，通常较长。
`path`（可选）为触发该错误的目标子路径；若为连接级错误或无子路径则省略。

| Code | Reason |
| -- | -- |
| llm_choice_empty | 在 OpenAI Completions 格式中，choices[0] 不存在。通常由输入内容触发安全限制而引发。 |
| llm_generate | 可能是 JSON 输出格式不规范，也可能是因为网络问题。但在 LLM 生成文本时发生。 |
| message_process | 难以归类的错误。但在 Message 处理时发生。 |
| invalid_path | client 传入的 `path` 不符合格式规范。 |
