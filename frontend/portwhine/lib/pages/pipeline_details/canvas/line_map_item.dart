import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/models/line_model.dart';

class LineMapItem extends StatelessWidget {
  const LineMapItem(this.model, {this.isConnecting = false, super.key});

  final LineModel model;
  final bool isConnecting;

  @override
  Widget build(BuildContext context) {
    // The temporary "connecting" line should be drawn on top of the whole
    // canvas so it reliably follows the cursor, independent of bounding-box
    // calculations and translation.
    if (isConnecting) {
      return RepaintBoundary(
        child: SizedBox.expand(
          child: CustomPaint(
            painter: BezierPainter(
              model: model,
              isConnecting: true,
              originX: 0,
              originY: 0,
            ),
          ),
        ),
      );
    }

    // Calculate the bounding box for the line
    final minX = math.min(model.startX, model.endX);
    final minY = math.min(model.startY, model.endY);
    final maxX = math.max(model.startX + nodeWidth, model.endX);
    final maxY = math.max(model.startY + nodeHeight, model.endY + nodeHeight);

    // Add padding for the curve and glow
    const padding = 50.0;

    final originX = minX - padding;
    final originY = minY - padding;
    final paintWidth = maxX - minX + padding * 2;
    final paintHeight = maxY - minY + padding * 2;

    return RepaintBoundary(
      child: Transform.translate(
        offset: Offset(originX, originY),
        child: SizedBox(
          width: paintWidth,
          height: paintHeight,
          child: CustomPaint(
            painter: BezierPainter(
              model: model,
              isConnecting: isConnecting,
              originX: originX,
              originY: originY,
            ),
            size: Size(paintWidth, paintHeight),
          ),
        ),
      ),
    );
  }
}

class BezierPainter extends CustomPainter {
  BezierPainter({
    required this.model,
    this.isConnecting = false,
    required this.originX,
    required this.originY,
  });

  final LineModel model;
  final bool isConnecting;
  final double originX;
  final double originY;

  @override
  void paint(Canvas canvas, Size size) {
    final startX = model.startX - originX;
    final startY = model.startY - originY;
    final endX = model.endX - originX;
    final endY = model.endY - originY;

    Offset n1 = Offset(
      startX + nodeWidth,
      startY + (nodeHeight / 2),
    );
    // For connecting lines, use the raw end position (mouse position)
    // For established connections, offset by nodeHeight/2 to center on the input port
    Offset n2 = Offset(
      endX,
      isConnecting ? endY : endY + (nodeHeight / 2),
    );

    // Calculate stretch based on distance for smoother curves
    final distance = (n2 - n1).distance;
    final stretch = math.min(distance * 0.4, 120.0);

    final path = Path()
      ..moveTo(n1.dx, n1.dy)
      ..cubicTo(
        n1.dx + stretch,
        n1.dy,
        n2.dx - stretch,
        n2.dy,
        n2.dx,
        n2.dy,
      );

    // Create gradient colors for the line
    final gradientColors = isConnecting
        ? [
            MyColors.prime.withValues(alpha: 0.8),
            MyColors.prime,
          ]
        : [
            const Color(0xFF6366F1), // Indigo
            const Color(0xFF8B5CF6), // Purple
          ];

    // Draw subtle glow effect
    final glowPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round
      ..color = gradientColors[0].withValues(alpha: 0.2);

    canvas.drawPath(path, glowPaint);

    // Draw main line with gradient
    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2
      ..strokeCap = StrokeCap.round
      ..shader = LinearGradient(
        colors: gradientColors,
      ).createShader(Rect.fromPoints(n1, n2));

    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(covariant BezierPainter oldDelegate) {
    return oldDelegate.model.startX != model.startX ||
        oldDelegate.model.startY != model.startY ||
        oldDelegate.model.endX != model.endX ||
        oldDelegate.model.endY != model.endY ||
        oldDelegate.isConnecting != isConnecting;
  }
}
