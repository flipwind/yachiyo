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
}
```

`message` 内是纯文本内容。
`time` 为 unix 时间戳。

## Runtime Message

指 runtime 向 client 发送的消息。  
同样在当前**仅支持文字**。

```
type: string = "runtime_message"
data {
    reply: bool
    message: string
    is_initiative: bool
    time: int64
}
```

`reply` 表示 runtime 是否自身决定回复。
`message` 内为消息。  
`is_initiative` 用于标示是否为 runtime 主动发出的消息。
`time` 为 unix 时间戳。

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
