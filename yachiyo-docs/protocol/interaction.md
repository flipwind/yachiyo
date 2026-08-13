# Interaction

当前阶段通常指 runtime 和 client 间的消息。

其中 `category` 标识为 `interaction`。

## Client Message

指 client 向 runtime 发送的消息。  
当前**仅支持文字**。

```
data {
    type: string = "client_message"
    message: string
}
```

`message` 内是纯文本内容。

## Runtime Message

指 runtime 向 client 发送的消息。  
同样在当前**仅支持文字**。

```
data {
    type: string = "runtime_message"
    message: string
    is_initiative: bool
}
```

`message` 内为消息。  
`is_initiative` 用于标示是否为 runtime 主动发出的消息。