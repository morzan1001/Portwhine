import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

TextStyle style({
  double size = 14,
  FontWeight weight = FontWeight.normal,
  Color color = CustomColors.textDark,
}) {
  return TextStyle(
    fontSize: size,
    fontWeight: weight,
    color: color,
  );
}

TextStyle headline1 = style(size: 32, weight: FontWeight.bold);
TextStyle headline2 = style(size: 28, weight: FontWeight.bold);
TextStyle headline3 = style(size: 24, weight: FontWeight.bold);
TextStyle headline4 = style(size: 20, weight: FontWeight.bold);
TextStyle headline5 = style(size: 18, weight: FontWeight.bold);
TextStyle headline6 = style(size: 16, weight: FontWeight.bold);

TextStyle subtitle1 = style(size: 16, weight: FontWeight.w500);
TextStyle subtitle2 = style(size: 14, weight: FontWeight.w500);

TextStyle bodyText1 = style(size: 16);
TextStyle bodyText2 = style(size: 14);

TextStyle button = style(size: 14, weight: FontWeight.w500, color: CustomColors.white);
TextStyle caption = style(size: 12);
TextStyle overline = style(size: 10);
