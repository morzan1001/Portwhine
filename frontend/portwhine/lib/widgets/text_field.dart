import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

class MyTextField extends StatelessWidget {
  final String hint;
  final String? label;
  final TextEditingController? controller;
  final TextInputType type;
  final int lines;
  final bool password;
  final bool expanded;
  final bool absorb;
  final bool isWhite;
  final List<TextInputFormatter>? inputFormatters;
  final ValueChanged<String>? onChanged;
  final Widget? prefixWidget, suffixWidget;
  final double radius;

  const MyTextField({
    super.key,
    this.label,
    required this.hint,
    this.controller,
    this.password = false,
    this.expanded = false,
    this.absorb = false,
    this.isWhite = false,
    this.type = TextInputType.text,
    this.lines = 1,
    this.inputFormatters,
    this.prefixWidget,
    this.suffixWidget,
    this.onChanged,
    this.radius = 12,
  });

  @override
  Widget build(BuildContext context) {
    final widget = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (label != null)
          Text(
            label!,
            overflow: TextOverflow.ellipsis,
            style: style(),
          ),
        if (label != null) const VerticalSpacer(6),
        Container(
          decoration: BoxDecoration(
            color: MyColors.grey,
            borderRadius: BorderRadius.circular(radius),
          ),
          child: AbsorbPointer(
            absorbing: absorb,
            child: TextField(
              onChanged: onChanged,
              textCapitalization: TextCapitalization.sentences,
              controller: controller,
              keyboardType: type,
              style: style(
                size: 16,
                color: MyColors.black,
                weight: FontWeight.w500,
              ),
              inputFormatters: inputFormatters,
              maxLines: lines,
              cursorColor: MyColors.black,
              decoration: InputDecoration(
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 18,
                  vertical: 12,
                ),
                hintText: hint,
                hintStyle: style(size: 15),
                prefixIcon: prefixWidget != null
                    ? Padding(
                        padding: const EdgeInsets.only(right: 12, left: 16),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [prefixWidget!],
                        ),
                      )
                    : null,
                suffixIcon: suffixWidget != null
                    ? Padding(
                        padding: const EdgeInsets.only(right: 12, left: 16),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [suffixWidget!],
                        ),
                      )
                    : null,
              ),
            ),
          ),
        ),
      ],
    );

    return expanded ? Expanded(child: widget) : widget;
  }
}
