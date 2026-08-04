import 'package:flutter/material.dart';

class ChatWidget extends StatefulWidget {
  const ChatWidget({super.key});

  @override
  State<ChatWidget> createState() => _ChatWidgetState();
}

class _ChatWidgetState extends State<ChatWidget> {
  String title = "Yachiyo Runtime";

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          Expanded(
            child: Card.filled(
              child: ListView.builder(
                itemBuilder: (context, index) {
                  return ListTile(title: Text("Item $index"));
                },
              ),
            ),
          ),
          Row(
              children: [
                Expanded(
                  child: Card.filled(
                    child: TextField(
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
                IconButton(onPressed: () => {}, icon: Icon(Icons.send_outlined)),
              ],
            ),
        ],
      ),
    );
  }
}
