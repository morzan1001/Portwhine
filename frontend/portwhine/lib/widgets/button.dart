import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

class Button extends StatelessWidget {
  final String text;
  final void Function()? onPressed;
  final Color buttonColor, textColor;
  final Color? borderColor;
  final double height, width, borderRadius, padding, textSize;
  final bool showAddIcon;
  final IconData? icon;

  const Button(
    this.text, {
    super.key,
    this.onPressed,
    this.buttonColor = CustomColors.prime,
    this.borderColor,
    this.textColor = CustomColors.white,
    this.height = 56,
    this.borderRadius = 8,
    this.padding = 4,
    this.textSize = 14,
    this.width = 0,
    this.showAddIcon = false,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(
          borderRadius,
        ),
        border: Border.all(color: borderColor ?? buttonColor),
      ),
      child: MaterialButton(
        hoverElevation: 0,
        highlightElevation: 0,
        elevation: 0,
        height: height,
        minWidth: width,
        onPressed: onPressed ?? () {},
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(
            borderRadius,
          ),
        ),
        color: buttonColor,
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: padding),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (showAddIcon)
                Icon(
                  Icons.add,
                  size: 20,
                  color: textColor,
                ),
              if (icon != null)
                Icon(
                  icon,
                  size: 20,
                  color: textColor,
                ),
              HorizontalSpacer(showAddIcon || icon != null ? 6 : 0),
              Flexible(
                child: Text(
                  text,
                  textAlign: TextAlign.center,
                  overflow: TextOverflow.ellipsis,
                  style: style(
                    color: textColor,
                    size: textSize,
                    weight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
