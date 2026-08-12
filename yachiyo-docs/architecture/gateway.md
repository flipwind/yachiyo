# Gateway

中间层，负责翻译 clients 和 runtime 之间的交互内容。

目前使用 JSON 作为临时通信协议。该部分不稳定，未来会有破坏性变更。

待 protocol 稳定后，gateway 内可能同时存在 gRPC 和 JSON 作为通信方式。