import 'dart:math';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:pancake/core/model/message.dart';

class ChatPane extends StatefulWidget {
  const ChatPane({super.key});

  @override
  State<StatefulWidget> createState() => _ChatPaneState();
}

class _ChatPaneState extends State<ChatPane> {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    return SafeArea(
      child: Column(
        children: [
          Padding(
            padding: EdgeInsetsGeometry.fromLTRB(24, 8, 16, 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              mainAxisSize: MainAxisSize.max,
              children: [
                Row(
                  spacing: 16.0,
                  children: [
                    Image.asset(
                      "assets/yachiyo_circle_icon.png",
                      height: 40,
                      width: 40,
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          "Yachiyo",
                          style: texts.titleLarge!.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          "iriririrooooooooha",
                          style: texts.titleSmall!.copyWith(
                            fontWeight: FontWeight.w300,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                IconButton(onPressed: () {}, icon: Icon(Icons.history)),
              ],
            ),
          ),
          Divider(thickness: 0.3),
          Expanded(
            child: Column(
              children: [
                Expanded(
                  child: ListView(
                    padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    children: [
                      for (
                        int i = 0, j = Random(i).nextInt(17);
                        i < 30;
                        i++, j = Random(i).nextInt(17)
                      )
                        if (i % 2 == 0)
                          _MessageContainerWidget(
                            message: UserMessage(
                              content: "content",
                              time: DateTime.now(),
                            ),
                          )
                        else
                          _MessageContainerWidget(
                            message: AssistantMessage(
                              content: "content",
                              time: DateTime.now(),
                              reply: (j % 3 == 0),
                              initiative: (j % 2 == 0),
                            ),
                          ),
                    ],
                  ),
                ),
                Card.outlined(
                  margin: EdgeInsetsGeometry.symmetric(
                    horizontal: 12,
                    vertical: 8,
                  ),
                  child: Padding(
                    padding: EdgeInsetsGeometry.fromLTRB(16, 4, 8, 4),
                    child: Row(
                      spacing: 8,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: TextField(
                            minLines: 1,
                            maxLines: 4,
                            decoration: InputDecoration(
                              hintText: "How about your day...",
                              border: InputBorder.none,
                            ),
                          ),
                        ),
                        Padding(
                          padding: EdgeInsetsGeometry.only(bottom: 4),
                          child: IconButton.filled(
                            onPressed: null,
                            style: IconButton.styleFrom(
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                            ),
                            icon: Icon(Icons.arrow_outward),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _EmptyMessagePlaceholder extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;

    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            margin: EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: colors.tertiaryContainer,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Padding(
              padding: EdgeInsetsGeometry.all(16),
              child: Icon(Icons.forum_outlined),
            ),
          ),
          Text("Chat", style: texts.titleMedium),
          Text("Some experience, or just say Hello.", style: texts.bodyMedium),
        ],
      ),
    );
  }
}

class _MessageContainerWidget extends StatelessWidget {
  final Message message;
  const _MessageContainerWidget({required this.message});

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final texts = Theme.of(context).textTheme;

    late Widget? profile;
    late bool? reply;
    late String nickname; // TODO: runtime name
    late DateTime time;
    late String content;
    late bool initiative;
    late bool isUser;

    switch (message) {
      case final AssistantMessage aMessage:
        profile = Image.asset(
          "assets/yachiyo_circle_icon.png",
          width: 32,
          height: 32,
        );
        reply = aMessage.reply;
        nickname = "Yachiyo";
        time = aMessage.time;
        content = aMessage.content;
        initiative = aMessage.initiative;
        isUser = false;
      case final UserMessage uMessage:
        profile = null;
        reply = null;
        nickname = "You";
        time = uMessage.time;
        content = uMessage.content;
        initiative = false;
        isUser = true;
    }

    if (isUser == false && reply == false) {
      return Padding(
        padding: EdgeInsetsGeometry.only(bottom: 12),
        child: Align(
          alignment: AlignmentGeometry.center,
          child: InkWell(
            onTap: () {},
            borderRadius: BorderRadius.circular(12),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              spacing: 8,
              children: [
                Icon(Icons.tips_and_updates_outlined, size: 12),
                Text(
                  "Chosen to not reply · ${DateFormat('kk:mm').format(time)}",
                  style: texts.bodySmall,
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Padding(
      padding: EdgeInsetsGeometry.only(bottom: 12),
      child: Row(
        mainAxisAlignment: isUser
            ? MainAxisAlignment.end
            : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.start,
        spacing: 8,
        children: [
          if (profile != null) profile,
          Column(
            crossAxisAlignment: isUser
                ? CrossAxisAlignment.end
                : CrossAxisAlignment.start,
            spacing: 4,
            children: [
              if (initiative)
                InkWell(
                  onTap: () {},
                  borderRadius: BorderRadius.circular(12),
                  child: Row(
                    spacing: 4,
                    children: [
                      Icon(Icons.auto_awesome, size: 10),
                      Text("Initiatively came.", style: texts.bodySmall),
                    ],
                  ),
                ),
              Container(
                padding: EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: colors.secondaryContainer,
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(isUser ? 12 : 4),
                    topRight: Radius.circular(isUser ? 4 : 12),
                    bottomLeft: Radius.circular(12),
                    bottomRight: Radius.circular(12),
                  ),
                ),
                child: Text(content),
              ),
              Text(
                "$nickname · ${DateFormat('kk:mm').format(time)}",
                style: texts.bodySmall,
              ),
            ],
          ),
        ],
      ),
    );
  }
}
