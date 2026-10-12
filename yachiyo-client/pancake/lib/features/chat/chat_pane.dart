import 'package:flutter/material.dart';

class ChatPane extends StatefulWidget {
  const ChatPane({super.key});

  @override
  State<StatefulWidget> createState() => _ChatPaneState();
}

class _ChatPaneState extends State<ChatPane> {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;
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
                  child: Center(
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
                        Text(
                          "Some experience, or just say Hello.",
                          style: texts.bodyMedium,
                        ),
                      ],
                    ),
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
