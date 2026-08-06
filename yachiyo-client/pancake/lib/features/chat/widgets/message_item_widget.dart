import 'package:flutter/material.dart';

class MessageItemWidget extends StatelessWidget {
  final String role;
  final String content;

  const MessageItemWidget({
    super.key,
    required this.role,
    required this.content,
  });

  @override
  Widget build(BuildContext context) {
    Widget headProfile(String role) {
      return Padding(
        padding: EdgeInsetsGeometry.all(16.0),
        child: Icon(Icons.person_pin),
      );
    }

    Widget messageContainer(String content) {
      return Flexible(
        child: Card(
          margin: EdgeInsetsGeometry.symmetric(vertical: 4.0),
          child: Padding(
            padding: EdgeInsetsGeometry.all(8.0),
            child: Text(content, softWrap: true),
          ),
        ),
      );
    }

    if (role == "assistant") {
      return Row(
        mainAxisAlignment: MainAxisAlignment.start,
        mainAxisSize: MainAxisSize.max,
        children: [headProfile(role), messageContainer(content)],
      );
    } else if (role == "user") {
      return Row(
        mainAxisAlignment: MainAxisAlignment.end,
        mainAxisSize: MainAxisSize.max,
        children: [messageContainer(content), headProfile(role)],
      );
    } else {
      return Row(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.max,
        children: [messageContainer(content)],
      );
    }
  }
}
