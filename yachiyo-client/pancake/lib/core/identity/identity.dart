import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class Identity {
  String clientType = "Client";
  String clientName = "Pancake!";
  String clientID;

  Identity(this.clientID);

  static Future<Identity> create() async {
    final pref = await SharedPreferences.getInstance();
    var clientID = pref.getString("client_id");

    if (clientID == null) {
      clientID = Uuid().v4();
      await pref.setString("client_id", clientID);
    }

    return Identity(clientID);
  }

  Future<void> refreshID() async {
    final pref = await SharedPreferences.getInstance();
    clientID = Uuid().v4();
    await pref.setString("client_id", clientID);
  }

  Future<void> changeID(String id) async {
    clientID = id;
    final pref = await SharedPreferences.getInstance();
    await pref.setString("client_id", id); 
  }
}