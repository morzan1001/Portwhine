import 'package:flutter/material.dart';

/// A reusable widget that provides hover and press scale animations.
/// Use this to avoid repeating MouseRegion + AnimatedContainer patterns.
class HoverScaleWidget extends StatefulWidget {
  const HoverScaleWidget({
    required this.child,
    this.onTap,
    this.hoverScale = 1.02,
    this.pressScale = 0.98,
    this.duration = const Duration(milliseconds: 200),
    this.cursor,
    super.key,
  });

  final Widget child;
  final VoidCallback? onTap;
  final double hoverScale;
  final double pressScale;
  final Duration duration;
  final MouseCursor? cursor;

  @override
  State<HoverScaleWidget> createState() => _HoverScaleWidgetState();
}

class _HoverScaleWidgetState extends State<HoverScaleWidget> {
  bool _isHovered = false;
  bool _isPressed = false;

  double get _scale {
    if (_isPressed) return widget.pressScale;
    if (_isHovered) return widget.hoverScale;
    return 1.0;
  }

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: widget.cursor ??
          (widget.onTap != null
              ? SystemMouseCursors.click
              : SystemMouseCursors.basic),
      child: GestureDetector(
        onTap: widget.onTap,
        onTapDown: (_) => setState(() => _isPressed = true),
        onTapUp: (_) => setState(() => _isPressed = false),
        onTapCancel: () => setState(() => _isPressed = false),
        child: AnimatedScale(
          scale: _scale,
          duration: widget.duration,
          curve: Curves.easeOutCubic,
          child: widget.child,
        ),
      ),
    );
  }
}
