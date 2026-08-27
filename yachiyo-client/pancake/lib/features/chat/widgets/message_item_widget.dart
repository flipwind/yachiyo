import 'package:flutter/material.dart';

class MessageItemWidget extends StatelessWidget {
  final String role;
  final String content;
  final double maxWidth;

  const MessageItemWidget({
    super.key,
    required this.role,
    required this.content,
    required this.maxWidth
  });

  @override
  Widget build(BuildContext context) {
    Widget headProfile(String role) {
      return Padding(
        padding: EdgeInsetsGeometry.all(16.0),
        child: Icon(Icons.person_pin),
      );
    }

    Widget messageContainer(String content, double maxWidth) {
      return ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: maxWidth,
        ),
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
        children: [headProfile(role), messageContainer(content, maxWidth)],
      );
    } else if (role == "user") {
      return Row(
        mainAxisAlignment: MainAxisAlignment.end,
        mainAxisSize: MainAxisSize.max,
        children: [messageContainer(content, maxWidth), headProfile(role)],
      );
    } else {
      return Row(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.max,
        children: [messageContainer(content, maxWidth)],
      );
    }
  }
}
