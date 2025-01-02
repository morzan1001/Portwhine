import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/widgets/loading_indicator.dart';

class MyIconButton extends StatelessWidget {
  final IconData icon;
  final void Function()? onTap;
  final Color buttonColor, iconColor;
  final double size, borderRadius, padding, iconSize;
  final bool showLoading;

  const MyIconButton(
    this.icon, {
    super.key,
    this.onTap,
    this.buttonColor = MyColors.lightGrey,
    this.iconColor = MyColors.black,
    this.size = 44,
    this.borderRadius = 12,
    this.padding = 4,
    this.iconSize = 22,
    this.showLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: !showLoading ? onTap : () {},
      borderRadius: BorderRadius.circular(borderRadius),
      child: Container(
        decoration: BoxDecoration(
          color: buttonColor,
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        height: size,
        width: size,
        child: Center(
          child: showLoading
              ? LoadingIndicator(color: iconColor, small: true)
              : Icon(icon, color: iconColor, size: iconSize),
        ),
      ),
    );
  }
}
