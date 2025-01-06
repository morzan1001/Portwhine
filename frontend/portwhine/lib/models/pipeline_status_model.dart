class PipelineStatusModel {
  final String id;
  final String status;

  PipelineStatusModel({
    this.id = '',
    this.status = '',
  });

  factory PipelineStatusModel.fromMap(Map<String, dynamic> map) {
    return PipelineStatusModel(
      id: map['id'] as String,
      status: map['status'] as String,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'status': status,
    };
  }

  PipelineStatusModel copyWith({
    String? id,
    String? status,
  }) {
    return PipelineStatusModel(
      id: id ?? this.id,
      status: status ?? this.status,
    );
  }
}
