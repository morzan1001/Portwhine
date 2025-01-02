class PipelineModel {
  final String id;
  final String name;

  PipelineModel({
    this.id = '',
    this.name = '',
  });

  // Factory constructor to create an instance from a map
  factory PipelineModel.fromMap(Map<String, dynamic> map) {
    return PipelineModel(
      id: map['id'] as String,
      name: map['name'] as String,
    );
  }

  // Method to convert the instance to a map
  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'name': name,
    };
  }

  // CopyWith method to create a modified copy of the instance
  PipelineModel copyWith({
    String? id,
    String? name,
  }) {
    return PipelineModel(
      id: id ?? this.id,
      name: name ?? this.name,
    );
  }

  @override
  String toString() => 'PipelineModel(id: $id, name: $name)';

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;

    return other is PipelineModel && other.id == id && other.name == name;
  }

  @override
  int get hashCode => id.hashCode ^ name.hashCode;
}
