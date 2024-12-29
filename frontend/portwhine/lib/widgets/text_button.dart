import 'package:flutter/material.dart';

class MyTextButton extends StatelessWidget {
  const MyTextButton(
    this.text, {
    this.color,
    this.onTap,
    super.key,
  });

  final String text;
  final Color? color;
  final void Function()? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return InkWell(
      onTap: onTap,
      child: Text(
        text,
        style: TextStyle(
          fontSize: 16,
          color: color ?? theme.colorScheme.secondary.withOpacity(0.75),
          decoration: TextDecoration.underline,
        ),
      ),
    );
  }
}
