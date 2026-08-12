# Client

从狭义上来说，client 可能指的是那些用于交互的应用程序；但当我们把它的义项拓展，那么在单 runtime 的情况下，似乎一切都可为 client。

而 Yachiyo Runtime 则将任何外部接入定义为 client；在交互过程中，必须有 gateway 作为中间层。

例如，广为人知的 onebot v11 可以与某些 IM 通信，我们的 gateway 在目前已经可以实现最简单的消息收发。在这个过程中，onebot v11 的实现端则可被视为 runtime 的 client 之一。

计划上，当 gateway 逐渐完善时，未来有相当一部分的设备可作为 client 接入 runtime。