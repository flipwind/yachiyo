import 'package:flutter/material.dart';
import 'package:pancake/core/model/yachiyo_model.dart';
import 'package:pancake/features/chat/widgets/message_item_widget.dart';
import 'package:provider/provider.dart';

class ChatWidget extends StatefulWidget {
  const ChatWidget({super.key});

  @override
  State<ChatWidget> createState() => _ChatWidgetState();
}

class _ChatWidgetState extends State<ChatWidget> {
  String title = "Yachiyo Runtime";
  String typeMessage = "";

  void sendMessage() {
    context.read<YachiyoModel>().sendMessage(typeMessage);
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          Expanded(
            child: Card.filled(
              child: Consumer<YachiyoModel>(
                builder: (context, model, child) {
                  return ListView.builder(
                    itemCount: model.messages.length,
                    itemBuilder: (context, index) {
                      final message = model.messages[index];
                      return MessageItemWidget(
                        role: message.role,
                        content: message.message,
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
                    onChanged: (value) {
                      typeMessage = value;
                    },
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
