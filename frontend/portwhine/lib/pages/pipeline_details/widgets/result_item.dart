import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';

class ResultItem extends StatelessWidget {
  const ResultItem({
    required this.result,
    required this.text,
    this.resultColor = CustomColors.sec,
    this.textColor = CustomColors.textLight,
    this.borderColor = CustomColors.grey,
    this.expanded = true,
    super.key,
  });

  final String result, text;
  final Color resultColor, textColor, borderColor;
  final bool expanded;

  @override
  Widget build(BuildContext context) {
    final widget = Container(
      width: 160,
      padding: const EdgeInsets.symmetric(
        vertical: 16,
      ),
      decoration: BoxDecoration(
        border: Border.all(color: borderColor, width: 0.5),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          Text(
            result,
            style: style(
              color: resultColor,
              weight: FontWeight.w600,
              size: 18,
            ),
          ),
          Text(
            text,
            style: style(
              color: textColor,
              weight: FontWeight.w400,
            ),
          ),
        ],
      ),
    );

    return expanded ? Expanded(child: widget) : widget;
  }
}
