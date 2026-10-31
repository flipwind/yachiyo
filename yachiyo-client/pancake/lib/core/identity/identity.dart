import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class Identity {
  String clientType = "Client";
  String clientName;
  String clientID;

  final SharedPreferencesAsync _pref = SharedPreferencesAsync();

  Identity(this.clientID, this.clientName);

  static Future<Identity> create() async {
    final _pref = SharedPreferencesAsync();
    String? clientID = await _pref.getString("client_id");
    String? clientName = await _pref.getString("client_name");

    if (clientID == null) {
      clientID = Uuid().v4();
      await _pref.setString("client_id", clientID);
    }

    if (clientName == null) {
      clientName = "Pancke!";
      await _pref.setString("client_name", "Pancake!");
    }

    return Identity(clientID, clientName);
  }

  Future<void> changeID(String id) async {
    clientID = id;
    await _pref.setString("client_id", id);
  }

  Future<void> changeName(String name) async {
    clientName = name;
    await _pref.setString("client_name", name);
  }
}
