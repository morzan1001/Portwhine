class PipelineModel {
  final String id;
  final String name;
  final String status;

  PipelineModel({
    this.id = '',
    this.name = '',
    this.status = '',
  });

  factory PipelineModel.fromMap(Map<String, dynamic> map) {
    return PipelineModel(
      id: map['id'] as String,
      name: map['name'] as String,
      status: map['status'] as String,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'name': name,
      'status': status,
    };
  }

  PipelineModel copyWith({
    String? id,
    String? name,
    String? status,
  }) {
    return PipelineModel(
      id: id ?? this.id,
      name: name ?? this.name,
      status: status ?? this.status,
    );
  }
}
