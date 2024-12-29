import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:portwhine/global/colors.dart';

void showToast(BuildContext context, String text, {Color? color}) {
  final toast = FToast().init(context);

  color = color ?? CustomColors.secDark;

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
          style: const TextStyle(
            color: Colors.white,
            fontSize: 14,
            fontWeight: FontWeight.w400,
          ),
        ),
      ),
    ),
  );
}
