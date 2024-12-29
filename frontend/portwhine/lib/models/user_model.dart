import 'package:portwhine/models/group_model.dart';

class UserModel {
  String name;
  String email;
  List<GroupModel> groups;
  DateTime? lastSignIn;
  bool online;

  UserModel({
    this.name = '',
    this.email = '',
    this.groups = const [],
    this.lastSignIn,
    this.online = false,
  });

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'email': email,
      'groups': groups.map((group) => group.toMap()).toList(),
      'lastSignIn': lastSignIn?.millisecondsSinceEpoch,
      'online': online,
    };
  }

  factory UserModel.fromMap(Map<String, dynamic> map) {
    return UserModel(
      name: map['name'],
      email: map['email'],
      groups: List<GroupModel>.from(
          map['groups'].map((x) => GroupModel.fromMap(x))),
      lastSignIn: map['lastSignIn'] != null
          ? DateTime.fromMillisecondsSinceEpoch(map['lastSignIn'])
          : null,
      online: map['online'],
    );
  }

  UserModel copyWith({
    String? name,
    String? email,
    List<GroupModel>? groups,
    DateTime? lastSignIn,
    bool? online,
  }) {
    return UserModel(
      name: name ?? this.name,
      email: email ?? this.email,
      groups: groups ?? this.groups,
      lastSignIn: lastSignIn ?? this.lastSignIn,
      online: online ?? this.online,
    );
  }
}
