import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/loading_indicator.dart';
import 'package:portwhine/widgets/spacer.dart';

class Button extends StatelessWidget {
  final String text;
  final void Function()? onTap;
  final Color buttonColor, textColor;
  final Color? borderColor;
  final double height, width, borderRadius, padding, textSize;
  final bool showAddIcon, showLoading;
  final IconData? icon;

  const Button(
    this.text, {
    super.key,
    this.onTap,
    this.buttonColor = MyColors.prime,
    this.textColor = MyColors.white,
    this.borderColor,
    this.height = 48,
    this.borderRadius = 12,
    this.padding = 24,
    this.textSize = 15,
    this.width = 0,
    this.showAddIcon = false,
    this.showLoading = false,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(borderRadius),
        border: borderColor != null
            ? Border.all(color: borderColor!, width: 0.5)
            : null,
      ),
      child: SizedBox(
        height: height,
        child: MaterialButton(
          hoverElevation: 0,
          highlightElevation: 0,
          elevation: 0,
          height: height,
          minWidth: width,
          onPressed: !showLoading ? onTap : () {},
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(borderRadius),
          ),
          padding: EdgeInsets.symmetric(horizontal: padding),
          color:
              !showLoading ? buttonColor : buttonColor.withValues(alpha: 0.8),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (!showLoading && showAddIcon)
                Icon(
                  Icons.add,
                  size: 18,
                  color: borderColor ?? textColor,
                ),
              if (!showLoading && icon != null)
                Icon(
                  icon!,
                  size: 20,
                  color: textColor,
                ),
              if (showLoading)
                LoadingIndicator(
                  color: borderColor ?? textColor,
                  small: true,
                ),
              HorizontalSpacer(
                showAddIcon || showLoading || (icon != null && text != '')
                    ? 6
                    : 0,
              ),
              if (!showLoading)
                Flexible(
                  child: Text(
                    text,
                    textAlign: TextAlign.center,
                    overflow: TextOverflow.ellipsis,
                    style: style(
                      color: borderColor ?? textColor,
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
