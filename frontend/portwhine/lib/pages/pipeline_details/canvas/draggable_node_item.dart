import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class DraggableNodeItem extends StatelessWidget {
  const DraggableNodeItem(this.node, {super.key});

  final NodeModel node;

  @override
  Widget build(BuildContext context) {
    final child = Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MyColors.white,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.drag_indicator,
            color: MyColors.black,
            size: 20,
          ),
          const HorizontalSpacer(4),
          Text(
            node.name,
            style: style(
              weight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );

    return Draggable<NodeModel>(
      data: node,
      feedback: Material(
        color: Colors.transparent,
        child: Transform.scale(
          scale: 1.05,
          child: Container(
            width: 240,
            decoration: BoxDecoration(
              color: MyColors.white,
              borderRadius: BorderRadius.circular(6),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.2),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                const Icon(
                  Icons.drag_indicator,
                  color: MyColors.black,
                  size: 20,
                ),
                const HorizontalSpacer(4),
                Text(
                  node.name,
                  style: style(
                    weight: FontWeight.w500,
                    color: MyColors.black,
                    decoration: TextDecoration.none,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
      child: child,
    );
  }
}
