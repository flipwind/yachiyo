<p style="font-size:60px" align="center">🥞</p>

<div align="center">

# Project Yachiyo
_✨ For the pancake. ✨_

</div>

## Introduce

Project Yachiyo is a project which **aims to build a life runtime**.  
Its final goal is to create a digital life, whether it's Yachiyo or any other character.

Since we have realized that LLMs alone are not the answer, I still believe that we can get infinitely closer to this goal.

Currently, the first version is **on preview**.  
The branch is still in active development.  
However, the first version will **inevitably be imperfect** and **may not be distinguished from similar products**. But it's important to take the first step.

Notice that the first version is **incomplete**.

## ✨ Quick Start

As you see, the first version is just **too simple**.  
As a result, these steps are also very simple.

> [!IMPORTANT]  
> Before starting, we hope that you have been already aware of **these risks**:  
> - Local first: We **didn't** do any authentication. As the result, please deploy it in trusted machine. Deploy it in open network is **highly not recommended**. Also, **exposed it directly to the LAN or the public network** are also not recommended.
> - Unsaved history: We think that our Yachiyo is currently unable to run a long time. So any storage now is in **in-memory**.  
> - Experimental protocol: Since our json protocol is under explore, it's **not suitable** for you to code or vibe a your own client.


Project Yachiyo consists of two parts, the server and the external clients.

### Server

1. Download the server package in **release page**.   
For non-provided platforms, try to compile according to the target platform. You can also just run `task server:build` to complete this step.
2. Edit `config.example.yaml` with your information, and **rename** it to `config.yaml`.
3. Run it.

### Client

Clients can be categorized into lots of kinds. Currently, there are only onebot, CLI and flutter clients which are built in.

> [!Note]  
> Currently, `Pancake!`, the flutter client is the only recommended client, for it has the most completed features.  
> The other clients in this monorepo, like onebot and CLI is too simple to be suitable for use.  

#### For flutter client (Recommend):

`Pancake!`, the flutter client, is designed to be a  cross-platform application.  
Most flutter-supported platforms, except `Android` and `Windows`, lack necessary tests and verifications, please notice.  
As the result, use `Pancake!` in platforms like `Linux` may have some unexpected problems.

`Pancake!` allows changing the connection address, monitoring runtime state and messaging with the runtime.  

In current period, `Pancake!` uses json over websocket to communicate with the runtime.  
A switch to gRPC is planned in the future.

1. Download the server package in **release page**.   
For non-provided platforms, try to compile according to the target platform. You can also just run `task pancake` to complete this step.
2. Run it.
3. Enter the server address field to monitor or message with the runtime.

#### For CLI client:

1. Compile according to the target platform.   
Or run `task cli:build` to compile it.
2. Run it. 

#### For onebot client:

This part is crude, and you may not hope it can work well.  

It don't need a specific client.   
The server serves as a Websocket Server, and the server have a gateway for translation.    
Use your onebot realization as a client, and use URL `ws://<host>:<port>/ws/onebot` to connect the server.

### About Prompt

Prompt for public is **a simplified version** of the origin.  
For some reason, the original prompt is private temporarily.

If you hope, you can provide your own system prompt.  

Any way, a **valid** prompt should have the following parts:

#### 1. Constrain the model's output.  
It may helps a lot by reading the code.  
Following is the **must** output json format.
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

> `Urgency` is a custom unit. And you can give its prompt to LLMs like this:
```md
Urgency is a custom unit that measures the level of importance.   
From lowest to highest, the six types are: Trivial, Low, Medium, High, Urgent, and Extreme.   
These terms will be given strictly according to this rule.   
If someone ask you to assign an Urgency, make sure to choose the appropriate type.
```

#### 2. Character's setup
As you wish.

## Documentation

The documents are stored in `yachiyo-docs`.

A simple development documentation site is available at [Project Yachiyo Docs](https://yachiyo.zako.ink).

> [!IMPORTANT] 
> Currently, Chinese is the only available documentation language.

## Contact

If you are interested in this project, discussions are welcome.  
You can contact me at `me[at]zako.ink` :>

Also, if you are a user, you can also join these groups to communicate with others. You are welcome to discuss your thoughts with us:  
- Telegram Group: [@project_yachiyo](https://t.me/project_yachiyo)
- QQ Group: [1126376305](https://qm.qq.com/q/u7w1xQ3a2k)

## License
This project is licensed under the [MIT License](LICENSE).