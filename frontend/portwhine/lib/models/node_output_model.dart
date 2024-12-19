class NodeOutputModel {
  String? outputs, name, type, value;

  NodeOutputModel({
    this.outputs,
    this.name,
    this.type,
    this.value,
  });

  Map<String, dynamic> toMap() {
    return {
      'outputs': outputs,
      'name': name,
      'type': type,
      'value': value,
    };
  }

  static NodeOutputModel fromMap(Map<String, dynamic> map) {
    return NodeOutputModel(
      outputs: map['outputs'],
      name: map['name'],
      type: map['type'],
      value: map['value'],
    );
  }

  NodeOutputModel copyWith({
    String? outputs,
    String? name,
    String? type,
    String? value,
  }) {
    return NodeOutputModel(
      outputs: outputs ?? this.outputs,
      name: name ?? this.name,
      type: type ?? this.type,
      value: value ?? this.value,
    );
  }
}
