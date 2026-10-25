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
    timetick: int64,
    energy: float64,
    affect: {
        octants: string,
        pleasure: float64,
        arousal: float64,
        dominance: float64,
        actives: List<OCCEmotion> = [
            {
                type: string,
                intensity: float64
            }
        ]
    }
}
```

`timetick` 表示 `Runtime State` 处理的时间，为 Unix 时间戳。
`energy` 表示精力值，在 `[0, 1]` 中。  
`affect` 表示情绪空间的各内容。  
- 其中：
    - `octants` 表示 <情绪点> 所处的象限类型。
    - `pleasure`，`arousal`，`dominance` 分别表示 <情绪点> 在空间中 x,y,z 三轴的位置，均处于 `[-1, 1]` 中。  
    - `actives` 表示空间中活跃的情绪**列表**。  
        - `type` 表示情绪的名称；
        - `intensity` 表示情绪的强度，在 `[0, 1]` 中。