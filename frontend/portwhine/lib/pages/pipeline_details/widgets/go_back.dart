import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

class GoBackButton extends StatelessWidget {
  const GoBackButton({
    this.onTap,
    this.color = CustomColors.white,
    super.key,
  });

  final Color color;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Icon(
            Icons.arrow_back,
            color: color,
            size: 18,
          ),
          const HorizontalSpacer(8),
          Text(
            'go back',
            style: style(
              color: color,
            ),
          )
        ],
      ),
    );
  }
}
