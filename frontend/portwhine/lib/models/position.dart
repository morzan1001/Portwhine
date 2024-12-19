class Position {
  final double x;
  final double y;

  const Position(this.x, this.y);

  Map<String, dynamic> toMap() {
    return {'x': x, 'y': y};
  }

  static Position fromMap(Map<String, dynamic> map) {
    return Position(map['x'], map['y']);
  }

  Position copyWith({double? x, double? y}) {
    return Position(x ?? this.x, y ?? this.y);
  }

  static const Position zero = Position(0, 0);
}
