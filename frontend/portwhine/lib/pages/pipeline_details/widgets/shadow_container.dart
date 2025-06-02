import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

class ShadowContainer extends StatelessWidget {
  const ShadowContainer({
    required this.child,
    this.padding = EdgeInsets.zero,
    this.onTap,
    super.key,
  });

  final Widget child;
  final EdgeInsets padding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 56,
        padding: padding,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          color: MyColors.white,
          boxShadow: [
            BoxShadow(
              blurRadius: 6,
              spreadRadius: 1,
              color: MyColors.black.withValues(alpha: 0.04),
            ),
          ],
        ),
        child: child,
      ),
    );
  }
}
