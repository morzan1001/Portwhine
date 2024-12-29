import 'package:portwhine/models/permission_model.dart';
import 'package:portwhine/models/user_model.dart';

class GroupModel {
  String name;
  List<UserModel> users;
  List<PermissionModel> permissions;

  static List<PermissionModel> defaultPermissions = [
    PermissionModel(name: 'Can edit users'),
    PermissionModel(name: 'Can delete users'),
    PermissionModel(name: 'Can edit pipelines'),
    PermissionModel(name: 'Can delete pipelines'),
    PermissionModel(name: 'Can create new accounts'),
    PermissionModel(name: 'REST API'),
  ];

  GroupModel({
    this.name = '',
    this.users = const [],
    List<PermissionModel>? permissions,
  }) : permissions = permissions ?? defaultPermissions;

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'users': users.map((user) => user.toMap()).toList(),
      'permissions':
          permissions.map((permission) => permission.toMap()).toList()
    };
  }

  factory GroupModel.fromMap(Map<String, dynamic> map) {
    return GroupModel(
      name: map['name'],
      users:
          List<UserModel>.from(map['users'].map((x) => UserModel.fromMap(x))),
      permissions: List<PermissionModel>.from(
          map['permissions'].map((x) => PermissionModel.fromMap(x))),
    );
  }

  GroupModel copyWith({
    String? name,
    List<UserModel>? users,
    List<PermissionModel>? permissions,
  }) {
    return GroupModel(
      name: name ?? this.name,
      users: users ?? this.users,
      permissions: permissions ?? this.permissions,
    );
  }
}

List<PermissionModel> defaultPermissions = [
  PermissionModel(name: 'Can edit users'),
  PermissionModel(name: 'Can delete users'),
  PermissionModel(name: 'Can edit pipelines'),
  PermissionModel(name: 'Can delete pipelines'),
  PermissionModel(name: 'Can create new accounts'),
  PermissionModel(name: 'REST API'),
];
