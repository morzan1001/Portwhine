import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

class ShadowContainer extends StatefulWidget {
  const ShadowContainer({
    required this.child,
    this.padding = EdgeInsets.zero,
    this.onTap,
    this.color,
    this.hoverColor,
    super.key,
  });

  final Widget child;
  final EdgeInsets padding;
  final VoidCallback? onTap;
  final Color? color;
  final Color? hoverColor;

  @override
  State<ShadowContainer> createState() => _ShadowContainerState();
}

class _ShadowContainerState extends State<ShadowContainer> {
  bool _isHovered = false;
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: widget.onTap != null
          ? SystemMouseCursors.click
          : SystemMouseCursors.basic,
      child: GestureDetector(
        onTap: widget.onTap,
        onTapDown: (_) => setState(() => _isPressed = true),
        onTapUp: (_) => setState(() => _isPressed = false),
        onTapCancel: () => setState(() => _isPressed = false),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          height: 48,
          padding: widget.padding,
          transform: Matrix4.diagonal3Values(
              _isPressed ? 0.97 : 1.0, _isPressed ? 0.97 : 1.0, 1.0),
          transformAlignment: Alignment.center,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            color: _isHovered
                ? (widget.hoverColor ??
                    const Color(0xFF6366F1).withValues(alpha: 0.08))
                : (widget.color ?? MyColors.white),
            border: Border.all(
              color: _isHovered
                  ? const Color(0xFF6366F1).withValues(alpha: 0.3)
                  : Colors.transparent,
              width: 1.5,
            ),
            boxShadow: [
              BoxShadow(
                blurRadius: _isHovered ? 16 : 10,
                spreadRadius: _isHovered ? 2 : 1,
                color: _isHovered
                    ? const Color(0xFF6366F1).withValues(alpha: 0.15)
                    : MyColors.black.withValues(alpha: 0.06),
                offset: Offset(0, _isHovered ? 6 : 3),
              ),
            ],
          ),
          child: widget.child,
        ),
      ),
    );
  }
}
