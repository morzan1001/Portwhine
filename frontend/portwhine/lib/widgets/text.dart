import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';

class SmallText extends StatelessWidget {
  const SmallText(
    this.text, {
    this.smaller = false,
    this.color = MyColors.textDarkGrey,
    super.key,
  });

  final String text;
  final bool smaller;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      style: style(color: color, size: smaller ? 12 : 14),
    );
  }
}

class Heading extends StatelessWidget {
  const Heading(
    this.text, {
    this.size = 16,
    this.color = MyColors.black,
    this.bold = false,
    super.key,
  });

  final String text;
  final double size;
  final Color color;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      style: style(
        size: size,
        color: color,
        weight: bold ? FontWeight.w600 : FontWeight.w500,
      ),
    );
  }
}

class BigHeading extends StatelessWidget {
  const BigHeading(
    this.text, {
    this.size = 20,
    this.color = MyColors.black,
    super.key,
  });

  final String text;
  final double size;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      style: style(size: size, color: color, weight: FontWeight.w600),
    );
  }
}
