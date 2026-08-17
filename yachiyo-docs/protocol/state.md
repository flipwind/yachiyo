# State

当前阶段多为状态更新。

其中 `category` 标识为 `state`。

## Runtime State

指 runtime 内部的 state 信息。  
当前是文本格式。

```
type: string = "runtime_state_request"
data {}
```

### Response

收到请求后，会将 state 信息转为字符串后发送。

```
type: string = "runtime_state"
data {
    state: string
}
```