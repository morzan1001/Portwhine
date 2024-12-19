class LineModel {
  double startX, startY, endX, endY;

  LineModel({
    required this.startX,
    required this.startY,
    required this.endX,
    required this.endY,
  });

  LineModel copyWith({
    double? startX,
    double? startY,
    double? endX,
    double? endY,
  }) {
    return LineModel(
      startX: startX ?? this.startX,
      startY: startY ?? this.startY,
      endX: endX ?? this.endX,
      endY: endY ?? this.endY,
    );
  }
}
