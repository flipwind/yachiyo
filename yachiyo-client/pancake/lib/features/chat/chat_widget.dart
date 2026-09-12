import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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
  final FocusNode focusNode = FocusNode();

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

  void clearMessages() {
    final provider = context.read<YachiyoProvider>();
    setState(() {
      provider.clearMessages();
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

                  return Column(
                    children: [
                      Expanded(
                        child: ListView.builder(
                          controller: listViewController,
                          itemCount: model.state.runtime.messages.length,
                          itemBuilder: (context, index) {
                            final message = model.state.runtime.messages[index];
                            return LayoutBuilder(
                              builder: (context, constraints) {
                                if (message.reply == false) {
                                  return NoReplyMessageItemWidget(
                                    character:
                                        model.state.runtime.runtimeName ??
                                        "Yachiyo",
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
                        ),
                      ),
                      Padding(
                        padding: EdgeInsetsGeometry.all(8.0),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.end,
                          children: [
                            IconButton(
                              tooltip: "Reload all messages",
                              onPressed: () {
                                // TODO: protocol support
                              },
                              icon: Icon(Icons.refresh_rounded),
                            ),
                            IconButton(
                              tooltip: "Clear all loaded messages",
                              onPressed: () {
                                clearMessages();
                              },
                              icon: Icon(Icons.delete_sweep_outlined),
                            ),
                          ],
                        ),
                      ),
                    ],
                  );
                },
              ),
            ),
          ),
          Card.outlined(
            child: Column(
              children: [
                Focus(
                  focusNode: focusNode,
                  onKeyEvent: (node, event) {
                    if (event is KeyDownEvent &&
                        event.logicalKey == LogicalKeyboardKey.enter) {
                      final isNewLine =
                          HardwareKeyboard.instance.isControlPressed ||
                          HardwareKeyboard.instance.isShiftPressed;

                      if (isNewLine) {
                        // ctrl + enter
                        // shift + enter
                        final value = textEditingController.value;
                        textEditingController.value = value.copyWith(
                          text: value.text.replaceRange(
                            value.selection.start,
                            value.selection.end,
                            "\n",
                          ),
                          selection: TextSelection.collapsed(
                            offset: value.selection.start + 1,
                          ),
                        );
                      } else {
                        // enter
                        sendMessage();
                      }

                      return KeyEventResult.handled;
                    }

                    return KeyEventResult.ignored;
                  },
                  child: TextField(
                    minLines: 1,
                    maxLines: 3,
                    keyboardType: TextInputType.multiline,
                    textInputAction: TextInputAction.newline,

                    controller: textEditingController,
                    decoration: InputDecoration(
                      hintText: 'Type message...',
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
                Padding(
                  padding: EdgeInsetsGeometry.fromLTRB(8.0, 0, 8.0, 8.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      IconButton.filled(
                        onPressed: () {
                          sendMessage();
                        },
                        icon: Icon(Icons.send_rounded),
                      ),
                    ],
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
