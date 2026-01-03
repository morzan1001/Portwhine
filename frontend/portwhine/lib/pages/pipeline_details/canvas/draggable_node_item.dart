import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/helpers.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class DraggableNodeItem extends StatefulWidget {
  const DraggableNodeItem(this.node, {super.key});

  final NodeModel node;

  @override
  State<DraggableNodeItem> createState() => _DraggableNodeItemState();
}

class _DraggableNodeItemState extends State<DraggableNodeItem>
    with SingleTickerProviderStateMixin {
  bool _isHovered = false;
  bool _isDragging = false;
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 150),
    );
    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 0.95,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeInOut));
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Color _getNodeTypeColor() => NodeHelper.getNodeColor(widget.node.name);

  IconData _getNodeIcon() => NodeHelper.getNodeIcon(widget.node.name);

  @override
  Widget build(BuildContext context) {
    final nodeColor = _getNodeTypeColor();

    final child = MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: SystemMouseCursors.grab,
      child: AnimatedBuilder(
        animation: _scaleAnimation,
        builder: (context, child) {
          return Transform.scale(
            scale: _isDragging ? 0.95 : (_isHovered ? 1.02 : 1.0),
            child: child,
          );
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          decoration: BoxDecoration(
            color: _isHovered
                ? nodeColor.withValues(alpha: 0.08)
                : MyColors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: _isHovered
                  ? nodeColor.withValues(alpha: 0.3)
                  : Colors.transparent,
              width: 1.5,
            ),
            boxShadow: _isHovered
                ? [
                    BoxShadow(
                      color: nodeColor.withValues(alpha: 0.15),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                    ),
                  ]
                : [
                    BoxShadow(
                      color: MyColors.black.withValues(alpha: 0.04),
                      blurRadius: 4,
                      offset: const Offset(0, 2),
                    ),
                  ],
          ),
          child: Row(
            children: [
              AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: nodeColor.withValues(alpha: _isHovered ? 0.15 : 0.1),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Icon(_getNodeIcon(), color: nodeColor, size: 16),
              ),
              const HorizontalSpacer(10),
              Expanded(
                child: Text(
                  widget.node.definition?.name ?? widget.node.name,
                  style: style(
                    weight: FontWeight.w500,
                    size: 13,
                    color: _isHovered ? nodeColor : MyColors.black,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              AnimatedOpacity(
                duration: const Duration(milliseconds: 200),
                opacity: _isHovered ? 1.0 : 0.3,
                child: Icon(
                  Icons.drag_indicator_rounded,
                  color: _isHovered ? nodeColor : MyColors.darkGrey,
                  size: 18,
                ),
              ),
            ],
          ),
        ),
      ),
    );

    return Draggable<NodeModel>(
      data: widget.node,
      onDragStarted: () => setState(() => _isDragging = true),
      onDragEnd: (_) => setState(() => _isDragging = false),
      feedback: Material(
        color: Colors.transparent,
        child: Transform.scale(
          scale: 1.05,
          child: Container(
            width: 260,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: MyColors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: nodeColor, width: 2),
              boxShadow: [
                BoxShadow(
                  color: nodeColor.withValues(alpha: 0.3),
                  blurRadius: 20,
                  offset: const Offset(0, 8),
                ),
                BoxShadow(
                  color: MyColors.black.withValues(alpha: 0.1),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: nodeColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(_getNodeIcon(), color: nodeColor, size: 18),
                ),
                const HorizontalSpacer(12),
                Expanded(
                  child: Text(
                    widget.node.definition?.name ?? widget.node.name,
                    style: style(
                      weight: FontWeight.w600,
                      size: 14,
                      color: MyColors.black,
                      decoration: TextDecoration.none,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
      childWhenDragging: AnimatedOpacity(
        duration: const Duration(milliseconds: 200),
        opacity: 0.4,
        child: child,
      ),
      child: child,
    );
  }
}
