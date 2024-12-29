import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
// import 'package:portwhine/blocs/password/obscure_password_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

class MyTextField extends StatelessWidget {
  final String label, hint;
  final TextEditingController controller;
  final TextInputType type;
  final int lines;
  final bool password;
  final bool expanded;
  final bool absorb;
  final List<TextInputFormatter>? inputFormatters;
  final Widget? icon;

  const MyTextField({
    super.key,
    required this.label,
    required this.hint,
    required this.controller,
    this.password = false,
    this.expanded = false,
    this.absorb = false,
    this.type = TextInputType.text,
    this.lines = 1,
    this.inputFormatters,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final widget = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: style(
            size: 14,
            weight: FontWeight.w500,
          ),
        ),
        const VerticalSpacer(6),
        Container(
          decoration: BoxDecoration(
            color: CustomColors.greyLight,
            borderRadius: BorderRadius.circular(8),
          ),
          padding: const EdgeInsets.symmetric(vertical: 2),
          child: AbsorbPointer(
            absorbing: absorb,
            child: TextField(
              textCapitalization: TextCapitalization.sentences,
              obscureText: password,
              controller: controller,
              keyboardType: type,
              style: const TextStyle(fontSize: 14),
              inputFormatters: inputFormatters,
              maxLines: lines,
              cursorColor: CustomColors.secDark,
              decoration: InputDecoration(
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 14,
                ),
                hintText: hint,
                hintStyle: const TextStyle(
                  fontSize: 14,
                ),
              ),
            ),
          ),
        ),
      ],
    );

    return expanded ? Expanded(child: widget) : widget;
  }
}
