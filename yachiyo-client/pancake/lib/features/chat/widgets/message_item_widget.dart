import 'package:flutter/material.dart';

class MessageItemWidget extends StatelessWidget {
  final String role;
  final String content;
  final double maxWidth;

  const MessageItemWidget({
    super.key,
    required this.role,
    required this.content,
    required this.maxWidth,
  });

  @override
  Widget build(BuildContext context) {
    Widget headProfile(String role) {
      return Padding(
        padding: EdgeInsetsGeometry.symmetric(horizontal: 8.0, vertical: 4.0),
        child: Icon(Icons.account_circle),
      );
    }

    Widget messageContainer(
      String content,
      double maxWidth, {
      bool leftSharp = false,
      bool rightSharp = false,
    }) {
      return ConstrainedBox(
        constraints: BoxConstraints(maxWidth: maxWidth),
        child: Card(
          margin: EdgeInsetsGeometry.symmetric(vertical: 4.0),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadiusGeometry.only(
              bottomLeft: Radius.circular(12.0),
              bottomRight: Radius.circular(12.0),
              topLeft: Radius.circular(leftSharp ? 0 : 12.0),
              topRight: Radius.circular(rightSharp ? 0 : 12.0),
            ),
          ),
          child: Padding(
            padding: EdgeInsetsGeometry.all(8.0),
            child: Text(content, softWrap: true),
          ),
        ),
      );
    }

    Widget messageRow() {
      if (role == "assistant") {
        return Row(
          mainAxisAlignment: MainAxisAlignment.start,
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.max,
          children: [
            headProfile(role),
            messageContainer(content, maxWidth, leftSharp: true),
          ],
        );
      } else if (role == "user") {
        return Row(
          mainAxisAlignment: MainAxisAlignment.end,
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.max,
          children: [
            messageContainer(content, maxWidth, rightSharp: true),
            headProfile(role),
          ],
        );
      } else {
        return Row(
          mainAxisAlignment: MainAxisAlignment.center,
          mainAxisSize: MainAxisSize.max,
          children: [messageContainer(content, maxWidth)],
        );
      }
    }

    return Padding(
      padding: EdgeInsetsGeometry.symmetric(vertical: 2.0),
      child: messageRow(),
    );
  }
}

class NoReplyMessageItemWidget extends StatelessWidget {
  final String character;

  const NoReplyMessageItemWidget({super.key, required this.character});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      mainAxisSize: MainAxisSize.max,
      children: [
        Card.outlined(
          child: Padding(
            padding: EdgeInsetsGeometry.all(8.0),
            child: Text("🍪 $character didn't reply."),
          ),
        ),
      ],
    );
  }
}
