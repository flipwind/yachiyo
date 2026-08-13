# Connection

通常关乎到 client 的整个生命周期。

其中 `category` 标识为 `connection`。

## Register

在与 runtime 交互的生命周期起始，必须要进行注册，以便 runtime 标识和分发其他内容。  
在**注册**后 client 才可以向 runtime 进行交互信息，否则会抛出错误。

> [!WARNING]
> 即，除 `Register` 消息外，所有消息均要求 client 已进行注册。

### 进行注册

此时，client 向 gateway 对应端口发送：
```
data {
    type: string = "register"
    client_type: string
    client_name: string
    client_id: string
}
```

- `client_type` 为设备类型，支持的设备类型：
    1. `IM`: 即时通讯实例，例如 Telegram，Discord 等。
    2. `Client`: 自定义的通讯实例，例如本项目的 Pancake。

- `client_name` 是一个字符串，该字段会被告知 LLM，建议增加其可读性。
- `client_id` 是该 client 的唯一标识符。请注意，相同 client 的多次连接应复用同一 `client_id`；推荐使用 uuid。  
    注意，同一 client 仅能同时拥有一个连接。当新的连接使用相同 `client_id` 注册时，gateway 将会拒绝或替换已有连接。

### 注册成功

而注册成功后，则 gateway 向 client 发送「注册成功」的消息。
```
data {
    type: string = "register_success"
    runtime_name: string
    runtime_version: string
}
```

其中 `runtime_name` 为 runtime 的名称；`runtime_version` 为连接 runtime 的版本号。

### 注册失败

如果注册失败，则 gateway 会抛出某种特定的错误类型。

```
data {
    type: string = "register_error"
    error_type: string
}
```

根据错误类型的不同，`error_type` 的值会在以下类型中产生：
- `client_info_error`: `client_type` 不在其字段限定中。建议检查字段内容的大小写和拼写。
- `client_conflict`: 注册设备冲突，表现为 `client_id` 存在重复但 `client_type` 不同。建议更换 `client_id`。

## Heartbeat

用于 runtime 和 client 互相告知存活信息。  
需要至少每 20 秒发送一次；对于时间波动，最多等待 30 秒，否则会触发 runtime 对 client 的自动下线。

必须先发送：
```
data {
    type: string = "heartbeat"
}
```

然后得到回复：
```
data {
    type: string = "heartbeat_respond"
}
```

## Offline

直接消息，对 runtime 发出 client 的主动下线告知。

```
data {
    type: string = "offline"
}
```