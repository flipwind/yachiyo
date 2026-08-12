# 浅谈架构

Project Yachiyo 旨在构筑一个类生命 runtime。在下面的内容里，我们同时把它命名为 `Yachiyo Runtime`。

与其他同方向的相似项目不同，它专注于 `runtime` 而非 `LLM`。正如 readme 内所写的，我们相信「纯 LLM 方案并不是最终答案」。

因此，一个 runtime 核心被构筑；一个全新的尝试被做出，以接近那个目标——创造，抑或延续一个生命。

## 组件们

大体上，`Yachiyo Runtime` 包含三大部分。

- 最重要的部分是 `yachiyo-runtime`，扮演「心脏」的角色。对于 runtime 的逻辑和状态，它们属于 runtime 的内容。
- 它的四肢则是 `yachiyo-client`。尽管现在它们被局限于小范围的平台，但是在未来的路线图上，它们最终会演变为可能的任何设备。
- 在上述两者之间的是 `yachiyo-gateway`。它们负责翻译 clients 和 runtime 的交互通信内容。

共同地，它们组成了 `Yachiyo Runtime` 的基础框架。
