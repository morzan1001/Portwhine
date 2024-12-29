import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:portwhine/global/colors.dart';

TextStyle style(
    {double? size, double? spacing, Color? color, FontWeight? weight}) {
  return GoogleFonts.inter(
    fontSize: size ?? 14,
    color: color ?? CustomColors.secDark,
    fontWeight: weight ?? FontWeight.normal,
    letterSpacing: spacing,
  );
}
