class PermissionModel {
  String name;
  bool status;

  PermissionModel({
    required this.name,
    this.status = false,
  });

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'status': status,
    };
  }

  factory PermissionModel.fromMap(Map<String, dynamic> map) {
    return PermissionModel(
      name: map['name'],
      status: map['status'],
    );
  }

  PermissionModel copyWith({
    String? name,
    bool? status,
  }) {
    return PermissionModel(
      name: name ?? this.name,
      status: status ?? this.status,
    );
  }
}
