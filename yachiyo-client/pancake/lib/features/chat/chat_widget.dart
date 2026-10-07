import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:pancake/core/model/message.dart';
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

  int previousMessageCount = 0;

  void sendMessage() {
    final provider = context.read<PancakeProvider>();
    final message = textEditingController.text;

    provider.sendMessage(message);
    setState(() {
      textEditingController.text = "";
    });
  }

  void clearMessages() {
    final provider = context.read<PancakeProvider>();
    setState(() {
      provider.clearMessages();
    });
  }

  void refreshMessageFromRuntime() {
    final provider = context.read<PancakeProvider>();
    provider.reloadMessagesFromRuntime();
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          Expanded(
            child: Card.filled(
              child: Consumer<PancakeProvider>(
                builder: (context, provider, child) {
                  final messageCount = provider.data.messages.length;

                  if (messageCount > previousMessageCount) {
                    WidgetsBinding.instance.addPostFrameCallback((_) {
                      if (listViewController.hasClients) {
                        listViewController.animateTo(
                          listViewController.position.maxScrollExtent,
                          duration: Duration(milliseconds: 400),
                          curve: Curves.easeOut,
                        );
                      }
                    });
                  }

                  previousMessageCount = messageCount;

                  return Column(
                    children: [
                      Expanded(
                        child: ListView.builder(
                          controller: listViewController,
                          itemCount: provider.data.messages.length,
                          itemBuilder: (context, index) {
                            final message = provider.data.messages[index];
                            return LayoutBuilder(
                              builder: (context, constraints) {
                                final maxWidth = constraints.maxWidth * 0.6;
                                switch (message) {
                                  case AssistantMessage():
                                    if (message.reply == false) {
                                      return NonReplyMessageItemWidget(
                                        character: provider.data.runtimeName,
                                      );
                                    } else {
                                      return AssistantItemWidget(
                                        content: message.content,
                                        maxWidth: maxWidth,
                                      );
                                    }
                                  case UserMessage():
                                    return UserItemWidget(
                                      content: message.content,
                                      maxWidth: maxWidth,
                                    );
                                  default:
                                    throw Exception(
                                      "Unknown message type: ${message.toString()}",
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
                                refreshMessageFromRuntime();
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
