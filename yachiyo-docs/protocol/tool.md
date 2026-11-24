# Tool

关于各 client 的能力，扩展 runtime 的接触面。

其中 `category` 为 `tool`。

## About Definition

扩展分为两种：

- `capability`，能力。仅作为字段上报，但**限定于** runtime 版本中支持的、固化了的字段，不支持自定义。
- `tool`，工具。提供 function 名，描述，以及需求字段；请求时会带上 requestID。

### capability

能力。**不支持任何的自定义**，一切字段仅以 runtime 对应的版本相定义。  
当前 runtime 支持的字段如下所示：

- message.text

在 `update_tools` 的 data 内的展示如：

```
capability: ["message.text", ...],
```

`capability`: 为一个 string 列表。字段限定于 runtime 规定的内容。

### tool

工具。支持自定义字段。提供时需要提供每个 tool 的具体内容：

- name：tool 的名称，需要唯一。限定于 32 个字符及以内，**仅允许** `a-z`, `A-Z`, `_`, `0-9` 字符。
- description：tool 的描述，提供给 llm。
- parameters：一只参数 map，**直接传给 llm 的 tool_call**，需要遵守 JSON Schema 规则。关键字部分请看[官方文档](https://json-schema.org/understanding-json-schema/keywords)。

单个展示如：

```
{
    "name": "get_user_location",
    "description": "get user currently location, returns a readable location string.",
    "parameters": {...}
}
```

## Update Tools

由 client **随时**上报自己的「能力」和「可调用工具」给 runtime，**全量更新**（覆盖）。  
当且仅当 client 上报后，runtime 中的 llm 组件才可将其「纳入囊中」。

上报时，发送：

```
type: string = "update_tools"
data {
    capability: ["message.text", ...],
    tools: [
        {
            "name": "get_user_location",
            "description": "get user currently location, returns a readable location string.",
            "parameters": {...}
        },
        ...
    ]
}
```

其中，

- `capability` 为一个 string 列表，包含 runtime 支持的任意字段名。不合法的 capability 会拒绝并在 feedback 中体现。
- `tools` 为一个 tool 列表，包含各个自定义的 `tool`。

### feedback

收到后，返回：

```
type: string = "update_tools_feedback"
data {
    status: "success" | "error",
    reject_capability: ["message.call"],
    reject_tools: {
        "get_user location": "name_error",
        ...
    }
}
```

其中，

- `status` 代表状态，string。只要存在错误即为 `error`，否则为 `success`。注意，未在返回中体现的 capability 或 tool 字段，被视为**已经生效**。
- `reject_capability` 表示**被拒绝部分的**类型，当且仅当 runtime 不支持时返回。检查拼写或 runtime 版本。
- `reject_tools` 为一个包含**被拒绝 tool** 的 map，其中 key 为 tool 的 name；value 为错误的原因，两者均为 string 格式。  
   请注意，tool 拒绝原因有以下字段：
  - `name_error`：表示未符合文档规定内容命名，比如 tool name 长度为 168。
  - `duplicate_name`：tool name 重复，请检查同一 client 是否有重复 tool 上报。

## ToolCall

调用对应工具。

runtime 发出请求：

```
type: string = "tool_call"
data {
    tool_name: string
    request_id: string
    payload: object
}
```

其中，

- `tool_name` 为 client 对 tools update 时的字段。
- `request_id` 为返回时需要**原样**返回的字段。
- `payload` 为 llm 调用函数时的参数 json payload；原样送达 client。请注意，由于 llm 的不确定性，**请对 payload 的各参数名称等进行校验**。

### result

由 client 返回给 runtime。
默认超时为 5s，超出该规定时间后的结果，或 request_id 不存在的返回结果将被**忽略**。

```
type: string = "tool_call_result"
data {
    request_id: string
    ok: bool
    result: string
}
```

其中，

- `request_id` 为 runtime 请求时的原样 id，**不可更改**。
- `ok` 为工具调用是否成功。
- `result` 为原样返回给 llm 的**可读结果**，无论结果是**错误**还是**成功**。
