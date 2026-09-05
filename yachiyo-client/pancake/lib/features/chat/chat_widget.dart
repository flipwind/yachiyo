import 'package:flutter/material.dart';
import 'package:pancake/core/model/message.dart';
import 'package:pancake/core/provider/yachiyo_provider.dart';
import 'package:pancake/features/chat/widgets/message_item_widget.dart';
import 'package:provider/provider.dart';

class ChatWidget extends StatefulWidget {
  const ChatWidget({super.key});

  @override
  State<ChatWidget> createState() => _ChatWidgetState();
}

class _ChatWidgetState extends State<ChatWidget> {
  String title = "Yachiyo Runtime";

  final TextEditingController textEditingController = TextEditingController();
  final ScrollController listViewController = ScrollController();

  void sendMessage() {
    final provider = context.read<YachiyoProvider>();
    final message = textEditingController.text;

    provider.state.runtime.messages.add(
      Message(role: "user", message: message, time: DateTime.now()),
    );
    provider.sendMessage(message);
    setState(() {
      textEditingController.text = "";
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          Expanded(
            child: Card.filled(
              child: Consumer<YachiyoProvider>(
                builder: (context, model, child) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (listViewController.hasClients) {
                      listViewController.animateTo(
                        listViewController.position.maxScrollExtent,
                        duration: Duration(milliseconds: 400),
                        curve: Curves.easeOut,
                      );
                    }
                  });

                  return ListView.builder(
                    controller: listViewController,
                    itemCount: model.state.runtime.messages.length,
                    itemBuilder: (context, index) {
                      final message = model.state.runtime.messages[index];
                      return LayoutBuilder(
                        builder: (context, constraints) {
                          if (message.reply == false) {
                            return NoReplyMessageItemWidget(
                              character:
                                  model.state.runtime.runtimeName ?? "Yachiyo",
                            );
                          } else {
                            return MessageItemWidget(
                              role: message.role,
                              content: message.message,
                              maxWidth: constraints.maxWidth * 0.6,
                            );
                          }
                        },
                      );
                    },
                  );
                },
              ),
            ),
          ),
          Row(
            children: [
              Expanded(
                child: Card.filled(
                  child: TextField(
                    controller: textEditingController,
                    onSubmitted: (_) {
                      sendMessage();
                    },
                    decoration: InputDecoration(
                      hintText: 'Typing Message...',
                      border: InputBorder.none,
                      enabledBorder: InputBorder.none,
                      focusedBorder: InputBorder.none,
                      contentPadding: EdgeInsets.symmetric(
                        horizontal: 16.0,
                        vertical: 12.0,
                      ),
                    ),
                  ),
                ),
              ),
              IconButton(
                onPressed: () {
                  sendMessage();
                },
                icon: Icon(Icons.send_outlined),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
