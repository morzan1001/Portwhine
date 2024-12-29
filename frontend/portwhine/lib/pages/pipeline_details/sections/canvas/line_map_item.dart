import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/models/line_model.dart';

class LineMapItem extends StatelessWidget {
  const LineMapItem(this.model, {super.key});

  final LineModel model;

  @override
  Widget build(BuildContext context) {
    return CustomPaint(painter: BezierPainter(model: model));
  }
}

class BezierPainter extends CustomPainter {
  BezierPainter({required this.model});

  final LineModel model;

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = CustomColors.greyDark
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    Offset n1 = Offset(
      model.startX + nodeWidth,
      model.startY + (nodeHeight / 2),
    );
    Offset n2 = Offset(
      model.endX,
      model.endY + (nodeHeight / 2),
    );
    const stretch = 70;

    final path = Path()
      ..moveTo(n1.dx, n1.dy)
      ..cubicTo(
        // first point
        n1.dx + stretch,
        n1.dy,
        // second point
        n2.dx - stretch,
        n2.dy,
        // final point
        n2.dx,
        n2.dy,
      );
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) {
    return true;
  }
}
