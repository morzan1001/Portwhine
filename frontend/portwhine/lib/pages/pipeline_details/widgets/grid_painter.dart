import 'package:flutter/material.dart';

class GridPainter extends CustomPainter {
  final double dotSpacing;
  final double dotRadius;
  final Color dotColor;
  final Color majorDotColor;
  final int majorGridInterval;

  const GridPainter({
    this.dotSpacing = 30.0,
    this.dotRadius = 1.5,
    this.dotColor = const Color(0xFFE0E0E0),
    this.majorDotColor = const Color(0xFFBDBDBD),
    this.majorGridInterval = 5,
  });

  @override
  void paint(Canvas canvas, Size size) {
    // Draw dot grid for a modern look
    final dotPaint = Paint()..style = PaintingStyle.fill;

    int xCount = 0;
    for (double x = 0; x < size.width; x += dotSpacing) {
      int yCount = 0;
      for (double y = 0; y < size.height; y += dotSpacing) {
        final isMajor =
            xCount % majorGridInterval == 0 && yCount % majorGridInterval == 0;

        dotPaint.color = isMajor ? majorDotColor : dotColor;
        final radius = isMajor ? dotRadius * 1.5 : dotRadius;

        canvas.drawCircle(Offset(x, y), radius, dotPaint);
        yCount++;
      }
      xCount++;
    }

    // Draw subtle cross lines at center
    final centerLinePaint = Paint()
      ..color = const Color(0xFFE8E8E8)
      ..strokeWidth = 1;

    final centerX = size.width / 2;
    final centerY = size.height / 2;

    // Horizontal center line
    canvas.drawLine(
      Offset(0, centerY),
      Offset(size.width, centerY),
      centerLinePaint,
    );

    // Vertical center line
    canvas.drawLine(
      Offset(centerX, 0),
      Offset(centerX, size.height),
      centerLinePaint,
    );
  }

  @override
  bool shouldRepaint(covariant GridPainter oldDelegate) {
    // Grid is static, never needs repaint
    return false;
  }
}
