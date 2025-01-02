import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

class LoadingIndicator extends StatelessWidget {
  const LoadingIndicator({this.color, this.small = false, super.key});

  final Color? color;
  final bool small;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Transform.scale(
        scale: small ? 0.5 : 0.8,
        child: CircularProgressIndicator(
          strokeWidth: small ? 4 : 2,
          valueColor: AlwaysStoppedAnimation<Color>(
            color ?? MyColors.prime,
          ),
        ),
      ),
    );
  }
}
