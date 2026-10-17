
<div align="center">
<img src="assets/picture/yachiyo_circle_icon.png" width="256" />

# Project Yachiyo
_✨ For the pancake. ✨_

[English README](README.en.md) · [MIT License](LICENSE)

</div>

## 简介

**Project Yachiyo** 是一个旨在构建「生命运行时」的项目。  
它的最终目标在于创建一个「数字生命」；当然，无论她是八千代还是其他的任何一个角色。

你可能会「ん?」，不过这并不奇怪。  
项目开发的原因，是因为我们认为仅靠 LLM 的方案不太能实现这个目标 —— 它可能并不是我们的最终答案。不过，我仍然坚信我们能够无限接近这个目标。

也因此，项目使用了一些**特别的设计**来向这个目标探索：  
例如提供状态；或者通过一些手段以触发主动消息，等等。其中有一些也可能是过往的经典设计？  
不过现在，其他未说明的大多数设想**还需要落地和验证**；由于一些个人经验的不足，这个过程可能有点缓慢。

> 非常欢迎探讨或者聊天！  
> 请跳转到 [#交流](#交流) 一节来与我联系 :>

当前，第一个版本仍处于预览期，主要面向开发人员和感兴趣的同学。它并不完美，并且可能无法区别于其他类似项目或 chatbot；但是第一步仍然是很重要的。

我们的项目仍在活跃开发中。请注意，第一个版本实际上并不完善。

## ✨ 快速开始

如你所见，第一个版本实在是太过简单。相应的，这些步骤**也许**也非常简单。（笑）

> [!IMPORTANT]  
> 在开始之前，我们得先知道这些风险：（免责声明）  
> - 本地优先。我们没做任何的认证与防护。因此，请将其部署在你的受信任的机器上。将其暴露于不受信任的局域网同样也是**不太推荐的**。也不太推荐将其**直接暴露于公网**。
> - 历史未落盘：由于 Yachiyo 目前并不是太适合长期运行，所以当前并没有做落盘。历史记录仅仅存在内存中，重启即丢失。
> - 实验性的协议：我们的 JSON 协议还在探索中，因此**不太推荐**你来通过编写或者 vibe 一个客户端。~~大量 token 持有者可以忽略这句话。~~

粗略划分的话，Project Yachiyo 包含两大部分。其一，服务端；其二，外部接入的客户端。

### 服务端
1. 下载 **Releases** 里的 server 程序。  
对于未提供相应平台的（真的会有吗），可以尝试本地编译看看。你也可以直接使用 `task server:build` 来完成这一步，构建的输出在 `/build` 里。
2. 使用你的配置，编辑相应的 `config.example.yaml`，并将其**重命名**为 `config.yaml`。
3. ✨ 尝试使用。

### 客户端

客户端们可以被分为许多种。当前内置的，只有 Onebot，CLI 以及 Flutter 客户端。

> [!Note]  
> 目前，Flutter 客户端 `Pancake!` 因着其最为完善的特性，是**唯一一个推荐的客户端**。  
> 在这个混合仓库的其他客户端，例如 Onebot 和 CLI，过于简陋，因此不太适合使用。

以下是内置客户端的简单说明。

#### 对于 Flutter 客户端（推荐）：

Flutter 客户端 `Pancake!` 被设计为跨平台应用，所以这也是使用 Flutter 的原因。  
**注意**！对于除了 `Android` 和 `Windows` 的其他 Flutter 支持平台，都缺少一些必要的测试与实机验证。  
也因此，在 `Linux` 之类的平台使用 `Pancake!` 可能会有些意想不到的问题。

`Pancake!` 允许您更改连接地址，监控 Runtime 的 State 属性，以及与 Runtime 进行交流。

当前阶段，`Pancake!` 使用基于 Websocket 的 JSON 协议进行双端通信。  
我们计划未来会使用 gRPC 之类的技术。

1. 在 Releases 里下载对应版本的 `Pancake!` 应用。  
对于未被提供的平台，你可以尝试着为你的平台自行编译。  
你也可以通过运行 `task pancake [-- BUILD_TARGET]` 来完成这一步。
    - `task pancake` 与 `task pancake:release` 默认编译 Windows 版本。  
    当然你也可以指定特定平台，比如对于 Android 可以用 `task pancake -- apk`，对于 Windows 则用 `task pancake -- windows`。
    - `[BUILD_TARGET]` 可以是 `aar`, `apk`, `web` 或其他 Flutter 支持平台的任何一种。运行 `flutter build` 来查阅其他支持的平台（这当然是对的，因为 task 任务肯定会使用 Flutter 的相关命令）。
2. 运行它。
3. 记得在应用内配置服务端地址。

#### 对于 CLI 平台：

1. 根据对应平台编译，或使用 `task cli:build`。
2. 运行。

#### 对于 Onebot 平台：

这部分太粗糙，并且你不能指望它很舒服地运行。

这并不是个特别的客户端。  
服务端作为一个 Websocket 服务，并且在连接过程中使用内置的 Gateway 进行协议翻译。  
在你的 Onebot 实现（例如 Napcat）中，将它们作为一个 Client；使用 `ws://<host>:<port>/ws/onebot` 以连接服务端。`port` 的内容取决于你的配置。

### 关于提示词的撰写

开源的 Prompt 是初版的简化版。  
由于这些那些的原因及顾虑，初版 Prompt 现在暂时是 private 的。  

如果你愿意，你当然可以撰写一份属于你自己的 system prompt。

但无论如何，一个可用的 prompt 需要含有以下的部分：

#### 1. 限制模型的输出

啃一下代码可能会有用。（我没说谁来啃.jpg）  
下面是模型输出必须遵循的 JSON 格式的**示意**，可以直接喂进去。

```json
{
    "reply": bool,
    "answer": string,
    "change": {
        "emotion": {
            "emotion": "happy" | "excited" | "sad" |... ,
            "result": Urgency
        },
        "state_delta": {
            "SocialDesire": -1 | 0 | 1,
            "Interest": -1 | 0 | 1
        }
    },
    "determination": {
        "should_open_a_new_session": bool,
        "should_wait_for_reply": bool
    },
    "note": string
}
```

> `Urgency` 是一个预设的单位。你可以这么告诉模型：
```md
Urgency 是一个预设的单位，用于测量重要程度。
六个子项从最弱到最强分别是 Trivial，Low，Medium，High，Urgent 和 Extreme。
这些术语需要严格按照规定输出；如果被需要输出一个 Urgency 单位，请选取合适的类型。
```

#### 2. 角色的设置。

请你自助。

## 文档

文档被存储在 `yachiyo-docs` 目录下。

可以访问一个简单的开发文档的站点 [Project Yachiyo Docs](https://yachiyo.zako.ink)。

## 交流

如果你对这个项目感兴趣，非常欢迎共同讨论！( •̀ ω •́ )✧  
你可以通过邮件 `me[at]zako.ink` 来联系我。

另外，如果你是一个使用者，你也可以加入我们的群组来交流讨论，无论是吹水还是你自己的想法：
- Telegram 群组：[@project_yachiyo](https://t.me/project_yachiyo)
- QQ 群组：[1126376305](https://qm.qq.com/q/u7w1xQ3a2k)

## 协议
Project Yachiyo 以 [MIT License](LICENSE) 的形式开源。

## 致谢

- **月见八千代** - 从 *[超かぐや姫！(Cosmic Princess Kaguya!)](https://www.cho-kaguyahime.com/)* 中姗姗走来的角色，是开发 Project Yachiyo 的楔子。
    - 默认的 system prompt 改编于其背景故事及其人格。
    - 默认的头像与项目图标，是在 ChatGPT 的协助下完成的 —— 基于角色原本的设计，并借鉴了 Grok Bot 的艺术风格。

    **我由衷感谢 *超かぐや姫！* 的创作者们，创作了一个激发该项目开发的角色。太伟大了。**
- **松饼图标** - 图标 [`pancake_icon.png`](/assets/picture/pancake.png) 取自以 MIT License 方式开源的 [Microsoft Fluent UI Emoji](https://github.com/microsoft/fluentui-emoji)。