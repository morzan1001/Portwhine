import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';

class MyTextButton extends StatelessWidget {
  const MyTextButton(
    this.text, {
    this.color = MyColors.prime,
    this.onTap,
    super.key,
  });

  final String text;
  final Color color;
  final void Function()? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Text(
        text,
        style: style(size: 16, color: color),
      ),
    );
  }
}
