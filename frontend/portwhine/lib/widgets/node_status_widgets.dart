import 'package:flutter/material.dart';
import 'package:portwhine/models/node_status.dart';
import 'package:portwhine/global/theme.dart';

/// Widget that displays the current status of a node with animated indicators.
class NodeStatusIndicator extends StatefulWidget {
  const NodeStatusIndicator({
    required this.status,
    this.size = 10,
    this.showPulse = true,
    super.key,
  });

  final NodeStatus status;
  final double size;
  final bool showPulse;

  @override
  State<NodeStatusIndicator> createState() => _NodeStatusIndicatorState();
}

class _NodeStatusIndicatorState extends State<NodeStatusIndicator>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    );
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.4).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
    _updateAnimation();
  }

  @override
  void didUpdateWidget(NodeStatusIndicator oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.status != widget.status) {
      _updateAnimation();
    }
  }

  void _updateAnimation() {
    if (widget.status.isActive && widget.showPulse) {
      _pulseController.repeat(reverse: true);
    } else {
      _pulseController.stop();
      _pulseController.reset();
    }
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  Color _getStatusColor(BuildContext context) {
    final colors = context.colors;
    switch (widget.status) {
      case NodeStatus.running:
        return colors.success;
      case NodeStatus.starting:
      case NodeStatus.restarting:
        return colors.warning;
      case NodeStatus.pending:
        return colors.info;
      case NodeStatus.completed:
        return colors.success;
      case NodeStatus.error:
      case NodeStatus.oomKilled:
      case NodeStatus.dead:
        return colors.error;
      case NodeStatus.stopped:
      case NodeStatus.paused:
        return colors.textSecondary;
      case NodeStatus.unknown:
        return colors.textTertiary;
    }
  }

  @override
  Widget build(BuildContext context) {
    final color = _getStatusColor(context);

    if (!widget.status.isActive || !widget.showPulse) {
      return _buildStaticIndicator(color);
    }

    return AnimatedBuilder(
      animation: _pulseAnimation,
      builder: (context, child) {
        return Stack(
          alignment: Alignment.center,
          children: [
            // Pulse ring
            Transform.scale(
              scale: _pulseAnimation.value,
              child: Container(
                width: widget.size,
                height: widget.size,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: color.withValues(alpha: 0.3 / _pulseAnimation.value),
                ),
              ),
            ),
            // Core dot
            child!,
          ],
        );
      },
      child: _buildStaticIndicator(color),
    );
  }

  Widget _buildStaticIndicator(Color color) {
    return Container(
      width: widget.size,
      height: widget.size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: color,
        boxShadow: [
          BoxShadow(
            color: color.withValues(alpha: 0.5),
            blurRadius: 6,
            spreadRadius: 1,
          ),
        ],
      ),
    );
  }
}

/// Widget that displays a status badge with icon and text.
class NodeStatusBadge extends StatelessWidget {
  const NodeStatusBadge({
    required this.status,
    this.compact = false,
    super.key,
  });

  final NodeStatus status;
  final bool compact;

  IconData _getStatusIcon() {
    switch (status) {
      case NodeStatus.running:
        return Icons.play_arrow_rounded;
      case NodeStatus.starting:
      case NodeStatus.restarting:
        return Icons.refresh_rounded;
      case NodeStatus.pending:
        return Icons.schedule_rounded;
      case NodeStatus.completed:
        return Icons.check_circle_rounded;
      case NodeStatus.error:
      case NodeStatus.oomKilled:
      case NodeStatus.dead:
        return Icons.error_rounded;
      case NodeStatus.stopped:
      case NodeStatus.paused:
        return Icons.stop_rounded;
      case NodeStatus.unknown:
        return Icons.help_outline_rounded;
    }
  }

  Color _getStatusColor(BuildContext context) {
    final colors = context.colors;
    switch (status) {
      case NodeStatus.running:
        return colors.success;
      case NodeStatus.starting:
      case NodeStatus.restarting:
        return colors.warning;
      case NodeStatus.pending:
        return colors.info;
      case NodeStatus.completed:
        return colors.success;
      case NodeStatus.error:
      case NodeStatus.oomKilled:
      case NodeStatus.dead:
        return colors.error;
      case NodeStatus.stopped:
      case NodeStatus.paused:
        return colors.textSecondary;
      case NodeStatus.unknown:
        return colors.textTertiary;
    }
  }

  @override
  Widget build(BuildContext context) {
    final color = _getStatusColor(context);

    if (compact) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Icon(
          _getStatusIcon(),
          size: 12,
          color: color,
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: color.withValues(alpha: 0.3),
          width: 1,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            _getStatusIcon(),
            size: 14,
            color: color,
          ),
          const SizedBox(width: 4),
          Text(
            status.displayName,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w500,
              color: color,
            ),
          ),
        ],
      ),
    );
  }
}

/// Animated border that shows activity around a node.
class NodeActivityBorder extends StatefulWidget {
  const NodeActivityBorder({
    required this.child,
    required this.status,
    this.isHovered = false,
    this.borderRadius = 16,
    super.key,
  });

  final Widget child;
  final NodeStatus status;
  final bool isHovered;
  final double borderRadius;

  @override
  State<NodeActivityBorder> createState() => _NodeActivityBorderState();
}

class _NodeActivityBorderState extends State<NodeActivityBorder>
    with SingleTickerProviderStateMixin {
  late AnimationController _rotationController;

  @override
  void initState() {
    super.initState();
    _rotationController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 3),
    );
    _updateAnimation();
  }

  @override
  void didUpdateWidget(NodeActivityBorder oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.status != widget.status) {
      _updateAnimation();
    }
  }

  void _updateAnimation() {
    if (widget.status.isActive) {
      _rotationController.repeat();
    } else {
      _rotationController.stop();
    }
  }

  @override
  void dispose() {
    _rotationController.dispose();
    super.dispose();
  }

  Color _getStatusColor(BuildContext context) {
    final colors = context.colors;
    if (widget.status.isError) return colors.error;
    if (widget.status.isActive) return colors.success;
    if (widget.status.isCompleted) return colors.info;
    return Colors.transparent;
  }

  @override
  Widget build(BuildContext context) {
    final showBorder = widget.status.isActive || widget.status.isError;
    final color = _getStatusColor(context);

    // Use Transform.scale for hover - more performant than AnimatedScale
    Widget child = widget.child;
    if (widget.isHovered) {
      child = Transform.scale(
        scale: 1.02,
        child: child,
      );
    }

    if (!showBorder) {
      return child;
    }

    return AnimatedBuilder(
      animation: _rotationController,
      builder: (context, animChild) {
        return Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(widget.borderRadius + 2),
            gradient: widget.status.isActive
                ? SweepGradient(
                    center: Alignment.center,
                    startAngle: 0,
                    endAngle: 3.14159 * 2,
                    transform: GradientRotation(
                        _rotationController.value * 3.14159 * 2),
                    colors: [
                      color.withValues(alpha: 0.0),
                      color.withValues(alpha: 0.5),
                      color.withValues(alpha: 0.0),
                    ],
                    stops: const [0.0, 0.5, 1.0],
                  )
                : null,
            border: widget.status.isError
                ? Border.all(color: color, width: 2)
                : null,
          ),
          padding: const EdgeInsets.all(2),
          child: animChild,
        );
      },
      child: child,
    );
  }
}
