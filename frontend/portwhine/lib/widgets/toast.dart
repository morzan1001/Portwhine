import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';

void showToast(
  BuildContext context,
  String text, {
  Color color = MyColors.prime,
}) {
  final toast = FToast().init(context);

  toast.showToast(
    positionedToastBuilder: (context, child, _) => Positioned(
      top: 48,
      left: 0,
      right: 0,
      child: child,
    ),
    toastDuration: const Duration(seconds: 3),
    child: ClipRRect(
      borderRadius: BorderRadius.circular(32),
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 12,
        ),
        color: color,
        child: Text(
          text,
          style: style(color: Colors.white),
        ),
      ),
    ),
  );
}
