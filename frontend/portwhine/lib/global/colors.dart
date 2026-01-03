import 'package:flutter/material.dart';

class MyColors {
  static const Color prime = Color(0xFFED1C24);

  static const Color grey = Color(0xFFF2F2F2);
  static const Color lightGrey = Color(0xFFF5F8F9);
  static const Color darkGrey = Color(0xFFBDBECC);
  static const Color textLightGrey = Color(0xFFA7A7AB);
  static const Color textDarkGrey = Color(0xFF686A79);
  static const Color background = Color(0xFFF5F8F9);
  static const Color border = Color(0xFFCBCBCB);

  static const Color white = Color(0xFFFFFFFF);
  static const Color black = Color(0xFF000000);

  static const Color green = Color(0xFF04CC24);
  static const Color red = Color(0xFFFF0000);

  // Theme colors (used throughout the app)
  static const Color indigo = Color(0xFF6366F1);
  static const Color purple = Color(0xFF8B5CF6);
  static const Color emerald = Color(0xFF10B981);
  static const Color amber = Color(0xFFF59E0B);
  static const Color rose = Color(0xFFEF4444);

  // Commonly used gradients
  static const LinearGradient primaryGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [indigo, purple],
  );

  /// Creates a subtle gradient for headers/backgrounds
  static LinearGradient subtleGradient(Color color) => LinearGradient(
        colors: [
          color.withValues(alpha: 0.1),
          color.withValues(alpha: 0.05),
        ],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      );
}
