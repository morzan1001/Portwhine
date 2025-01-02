import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodesList extends StatefulWidget {
  const NodesList({super.key});

  @override
  State<NodesList> createState() => _NodesListState();
}

class _NodesListState extends State<NodesList> {
  final nodes = ['Domain', 'Nmap', 'Nikto', 'Node'];

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 280,
      padding: const EdgeInsets.symmetric(
        horizontal: 20,
      ),
      decoration: BoxDecoration(
        color: MyColors.darkGrey.withOpacity(0.3),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: MyColors.darkGrey, width: 0.5),
        boxShadow: [
          BoxShadow(
            blurRadius: 6,
            spreadRadius: 1,
            color: MyColors.black.withOpacity(0.04),
          ),
        ],
      ),
      child: Theme(
        data: ThemeData(
          dividerColor: Colors.transparent,
        ),
        child: ExpansionTile(
          initiallyExpanded: true,
          tilePadding: EdgeInsets.zero,
          childrenPadding: const EdgeInsets.only(bottom: 20),
          title: Row(
            children: [
              Expanded(
                child: Text(
                  'NODES',
                  style: style(
                    spacing: 4,
                    color: MyColors.textLightGrey,
                    size: 12,
                  ),
                ),
              ),
              Row(
                children: [
                  const Icon(
                    Icons.add,
                    color: MyColors.black,
                    size: 20,
                  ),
                  const HorizontalSpacer(2),
                  Text(
                    'Add',
                    style: style(
                      weight: FontWeight.w500,
                    ),
                  ),
                ],
              )
            ],
          ),
          children: [
            ListView.separated(
              separatorBuilder: (a, b) => const VerticalSpacer(12),
              itemCount: nodes.length,
              shrinkWrap: true,
              itemBuilder: (_, i) {
                return Draggable<NodeModel>(
                  dragAnchorStrategy: pointerDragAnchorStrategy,
                  data: NodeModel(name: nodes[i]),
                  feedback: NodeSelectionItem(nodes[i]),
                  child: NodeSelectionItem(nodes[i]),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

class NodeSelectionItem extends StatelessWidget {
  const NodeSelectionItem(this.name, {super.key});

  final String name;

  @override
  Widget build(BuildContext context) {
    return Material(
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 12,
        ),
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
              name,
              style: style(
                weight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
