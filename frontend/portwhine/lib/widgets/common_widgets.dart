import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';

/// A container with an icon inside, commonly used for headers and list items.
/// Provides consistent styling across the app.
class IconContainer extends StatelessWidget {
  const IconContainer({
    required this.icon,
    required this.color,
    this.size = 18,
    this.padding = 8,
    this.borderRadius = 8,
    super.key,
  });

  final IconData icon;
  final Color color;
  final double size;
  final double padding;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(padding),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(borderRadius),
      ),
      child: Icon(
        icon,
        color: color,
        size: size,
      ),
    );
  }
}

/// A status indicator dot with optional glow effect.
class StatusDot extends StatelessWidget {
  const StatusDot({
    required this.color,
    this.size = 8,
    this.showGlow = false,
    super.key,
  });

  final Color color;
  final double size;
  final bool showGlow;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: color,
        boxShadow: showGlow
            ? [
                BoxShadow(
                  color: color.withValues(alpha: 0.5),
                  blurRadius: 6,
                  spreadRadius: 1,
                ),
              ]
            : null,
      ),
    );
  }
}

/// A card with consistent shadow and hover effects.
class HoverCard extends StatefulWidget {
  const HoverCard({
    required this.child,
    this.onTap,
    this.padding = const EdgeInsets.all(16),
    this.borderRadius = 16,
    this.color,
    this.hoverBorderColor,
    super.key,
  });

  final Widget child;
  final VoidCallback? onTap;
  final EdgeInsets padding;
  final double borderRadius;
  final Color? color;
  final Color? hoverBorderColor;

  @override
  State<HoverCard> createState() => _HoverCardState();
}

class _HoverCardState extends State<HoverCard> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final borderColor = widget.hoverBorderColor ?? MyColors.indigo;

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOutCubic,
        transform: Matrix4.diagonal3Values(
            _isHovered ? 1.01 : 1.0, _isHovered ? 1.01 : 1.0, 1.0),
        transformAlignment: Alignment.center,
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            borderRadius: BorderRadius.circular(widget.borderRadius),
            onTap: widget.onTap,
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              padding: widget.padding,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(widget.borderRadius),
                color: widget.color ?? MyColors.white,
                border: Border.all(
                  color: _isHovered
                      ? borderColor.withValues(alpha: 0.5)
                      : Colors.transparent,
                  width: 2,
                ),
                boxShadow: [
                  BoxShadow(
                    color: _isHovered
                        ? borderColor.withValues(alpha: 0.15)
                        : MyColors.black.withValues(alpha: 0.05),
                    blurRadius: _isHovered ? 20 : 8,
                    spreadRadius: _isHovered ? 1 : 0,
                    offset: Offset(0, _isHovered ? 6 : 2),
                  ),
                ],
              ),
              child: widget.child,
            ),
          ),
        ),
      ),
    );
  }
}

/// A header container with gradient background.
class GradientHeader extends StatelessWidget {
  const GradientHeader({
    required this.child,
    required this.color,
    this.padding = const EdgeInsets.all(24),
    this.borderRadius,
    super.key,
  });

  final Widget child;
  final Color color;
  final EdgeInsets padding;
  final BorderRadius? borderRadius;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: padding,
      decoration: BoxDecoration(
        gradient: MyColors.subtleGradient(color),
        borderRadius: borderRadius,
      ),
      child: child,
    );
  }
}

/// A simple hover button with icon and optional label.
class HoverIconButton extends StatefulWidget {
  const HoverIconButton({
    required this.icon,
    required this.onTap,
    this.color,
    this.size = 36,
    this.iconSize = 20,
    this.tooltip,
    this.enabled = true,
    super.key,
  });

  final IconData icon;
  final VoidCallback onTap;
  final Color? color;
  final double size;
  final double iconSize;
  final String? tooltip;
  final bool enabled;

  @override
  State<HoverIconButton> createState() => _HoverIconButtonState();
}

class _HoverIconButtonState extends State<HoverIconButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final color = widget.color ?? MyColors.indigo;

    Widget button = MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: widget.enabled
          ? SystemMouseCursors.click
          : SystemMouseCursors.forbidden,
      child: GestureDetector(
        onTap: widget.enabled ? widget.onTap : null,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          width: widget.size,
          height: widget.size,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            color: _isHovered && widget.enabled
                ? color.withValues(alpha: 0.15)
                : Colors.transparent,
          ),
          child: Center(
            child: Icon(
              widget.icon,
              color: widget.enabled
                  ? (_isHovered ? color : color.withValues(alpha: 0.7))
                  : MyColors.darkGrey,
              size: widget.iconSize,
            ),
          ),
        ),
      ),
    );

    if (widget.tooltip != null) {
      button = Tooltip(message: widget.tooltip!, child: button);
    }

    return button;
  }
}
