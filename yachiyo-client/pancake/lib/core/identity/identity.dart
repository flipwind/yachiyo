import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class Identity {
  String clientType = "Client";
  String clientName;
  String clientID;

  Identity(this.clientID, this.clientName);

  static Future<Identity> create() async {
    final pref = await SharedPreferences.getInstance();
    var clientID = pref.getString("client_id");
    var clientName = pref.getString("client_name");

    if (clientID == null) {
      clientID = Uuid().v4();
      await pref.setString("client_id", clientID);
    }

    if (clientName == null) {
      clientName = "Pancke!";
      await pref.setString("client_name", "Pancake!");
    }

    return Identity(clientID, clientName);
  }

  Future<void> changeID(String id) async {
    clientID = id;
    final pref = await SharedPreferences.getInstance();
    await pref.setString("client_id", id);
  }

  Future<void> changeName(String name) async {
    clientName = name;
    final pref = await SharedPreferences.getInstance();
    await pref.setString("client_name", name);
  }
}
