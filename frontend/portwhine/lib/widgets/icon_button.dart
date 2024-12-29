import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

class CustomIconButton extends StatelessWidget {
  final IconData icon;
  final void Function()? onPressed;
  final Color? buttonColor, iconColor;
  final double height, width, borderRadius, padding, iconSize;

  const CustomIconButton(
    this.icon, {
    super.key,
    this.onPressed,
    this.buttonColor,
    this.iconColor,
    this.height = 52,
    this.borderRadius = 8,
    this.padding = 4,
    this.iconSize = 18,
    this.width = 48,
  });

  @override
  Widget build(BuildContext context) {
    return MaterialButton(
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
      color: buttonColor ?? CustomColors.prime,
      child: Icon(
        icon,
        color: iconColor ?? Colors.white,
        size: iconSize,
      ),
    );
  }
}
