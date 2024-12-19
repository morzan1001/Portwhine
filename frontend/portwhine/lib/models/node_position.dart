class NodePosition {
  double x;
  double y;

  NodePosition({
    required this.x,
    required this.y,
  });

  Map<String, dynamic> toMap() {
    return {'x': x, 'y': y};
  }

  static NodePosition fromMap(Map<String, dynamic> map) {
    return NodePosition(x: map['x'], y: map['y']);
  }

  NodePosition copyWith({double? x, double? y}) {
    return NodePosition(x: x ?? this.x, y: y ?? this.y);
  }
}
