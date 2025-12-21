import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';
import 'package:portwhine/models/position.dart';
import 'package:portwhine/pages/pipeline_details/node_details.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeMapItem extends StatelessWidget {
  const NodeMapItem(this.model, {super.key});

  final NodeModel model;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () async {
        final canvas = BlocProvider.of<CanvasCubit>(context).state;

        final currentX = model.position!.x;
        final desiredX =
            (((width(context) - nodeWidth) / 2) - canvas.position.x) /
                canvas.zoom;
        final translateX = (desiredX - currentX).roundToDouble();

        final currentY = model.position!.y;
        final desiredY = (100 - canvas.position.y) / canvas.zoom;
        final translateY = (desiredY - currentY).roundToDouble();

        BlocProvider.of<CanvasCubit>(context).changePosition(
          Position(translateX, translateY),
        );

        if (translateX != 0 || translateY != 0) {
          await Future.delayed(const Duration(milliseconds: 500));
        }

        BlocProvider.of<SelectedNodeCubit>(context).setNode(model);
        await showNodeDetailsDialog(context, model);
        BlocProvider.of<SelectedNodeCubit>(context).removeNode();
      },
      onPanUpdate: (details) {
        final canvas = BlocProvider.of<CanvasCubit>(context).state;

        // Use delta for smoother movement relative to current position
        // Adjust delta by zoom level
        final dx = details.delta.dx / canvas.zoom;
        final dy = details.delta.dy / canvas.zoom;

        final newX = model.position!.x + dx;
        final newY = model.position!.y + dy;

        BlocProvider.of<NodesCubit>(context).moveNode(
          model.id,
          NodePosition(x: newX, y: newY),
        );

        BlocProvider.of<LinesCubit>(context).updateLines(
          BlocProvider.of<NodesCubit>(context).state,
        );
      },
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          Container(
            width: nodeWidth,
            height: nodeHeight,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              color: MyColors.white,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  model.name,
                  style: style(
                    color: MyColors.black,
                    weight: FontWeight.w600,
                  ),
                ),
                const VerticalSpacer(12),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: ListView.separated(
                        separatorBuilder: (a, b) => const VerticalSpacer(12),
                        itemCount: model.inputs.length,
                        shrinkWrap: true,
                        itemBuilder: (_, i) {
                          return InputOutputItem(
                            model.inputs.entries.toList()[i],
                          );
                        },
                      ),
                    ),
                    Expanded(
                      child: ListView.separated(
                        separatorBuilder: (a, b) => const VerticalSpacer(12),
                        itemCount: model.outputs.length,
                        shrinkWrap: true,
                        itemBuilder: (_, i) {
                          return InputOutputItem(
                            model.outputs.entries.toList()[i],
                          );
                        },
                      ),
                    ),
                  ],
                )
              ],
            ),
          ),
          Positioned(
            right: -8,
            top: 0,
            bottom: 0,
            child: NodeOutput(model),
          ),
          Positioned(
            left: -8,
            top: 0,
            bottom: 0,
            child: NodeInput(model),
          ),
        ],
      ),
    );
  }
}

class ConnectorWidget extends StatelessWidget {
  const ConnectorWidget({super.key});

  @override
  Widget build(BuildContext context) {
    final canvas = BlocProvider.of<CanvasCubit>(context).state;
    final zoom = canvas.zoom;
    // final zoom = 1.0;
    final size = 16 * zoom;
    final innerSize = 8 * zoom;

    return SizedBox(
      height: nodeHeight * zoom,
      child: Container(
        width: size,
        height: size,
        decoration: const BoxDecoration(
          shape: BoxShape.circle,
          color: MyColors.white,
        ),
        child: Center(
          child: Container(
            width: innerSize,
            height: innerSize,
            decoration: const BoxDecoration(
              shape: BoxShape.circle,
              color: MyColors.darkGrey,
            ),
          ),
        ),
      ),
    );
  }
}

class NodeInput extends StatelessWidget {
  const NodeInput(this.model, {super.key});

  final NodeModel model;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.grabbing,
      child: DragTarget<Map>(
        builder: (a, b, c) => const ConnectorWidget(),
        // removed onWillAccept
        onAcceptWithDetails: (details) {
          final outputNode = details.data['node'];
          final inputNode = model;

          BlocProvider.of<NodesCubit>(context).addConnection(
            outputNode,
            inputNode,
          );

          BlocProvider.of<LinesCubit>(context).updateLines(
            BlocProvider.of<NodesCubit>(context).state,
          );
        },
      ),
    );
  }
}

class NodeOutput extends StatelessWidget {
  const NodeOutput(this.model, {super.key});

  final NodeModel model;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.grab,
      child: Draggable<Map>(
        onDragStarted: () {
          BlocProvider.of<ConnectingLineCubit>(context).init(
            model.position!.x,
            model.position!.y,
          );
        },
        onDragUpdate: (details) {
          final canvas = BlocProvider.of<CanvasCubit>(context).state;
          BlocProvider.of<ConnectingLineCubit>(context).updateLine(
            (details.globalPosition.dx - canvas.position.x) / canvas.zoom,
            (details.globalPosition.dy - canvas.position.y - 90) / canvas.zoom -
                76,
          );
        },
        onDragEnd: (details) {
          BlocProvider.of<ConnectingLineCubit>(context).remove();
        },
        data: {'node': model},
        feedback: Container(),
        child: const ConnectorWidget(),
      ),
    );
  }
}

class InputOutputItem extends StatelessWidget {
  const InputOutputItem(this.entry, {super.key});

  final MapEntry entry;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          entry.key,
          style: style(
            size: 12,
            spacing: 2,
            color: MyColors.textLightGrey,
          ),
        ),
        const VerticalSpacer(2),
        Text(
          '${entry.value}',
          style: style(
            size: 14,
            color: MyColors.black,
            weight: FontWeight.w600,
          ),
        ),
      ],
    );
  }
}
