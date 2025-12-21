import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';
import 'package:portwhine/pages/pipeline_details/canvas/line_map_item.dart';
import 'package:portwhine/pages/pipeline_details/canvas/node_map_item.dart';
import 'package:portwhine/pages/pipeline_details/widgets/grid_painter.dart';

class PipelineCanvas extends StatefulWidget {
  const PipelineCanvas({super.key});

  @override
  State<PipelineCanvas> createState() => _PipelineCanvasState();
}

class _PipelineCanvasState extends State<PipelineCanvas> {
  final GlobalKey _canvasKey = GlobalKey();

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CanvasCubit, CanvasModel>(
      builder: (context, state) {
        final controller = state.controller;

        return InteractiveViewer(
          constrained: false,
          transformationController: controller,
          boundaryMargin: const EdgeInsets.all(double.infinity),
          minScale: 0.1,
          maxScale: 5.0,
          onInteractionEnd: (_) {
            final cubit = BlocProvider.of<CanvasCubit>(context);
            final position = controller!.value.getTranslation();
            final scale = controller.value.getMaxScaleOnAxis();
            cubit.setPosition(position.x, position.y);
            cubit.setZoom(scale);
          },
          child: SizedBox(
            width: 5000,
            height: 5000,
            child: BlocBuilder<LinesCubit, List<LineModel>>(
              builder: (context, linesState) {
                return BlocBuilder<NodesCubit, List<NodeModel>>(
                  builder: (context, nodesState) {
                    return Stack(
                      key: _canvasKey,
                      children: [
                        // Grid background for better visualization
                        Positioned.fill(
                          child: CustomPaint(
                            painter: GridPainter(),
                          ),
                        ),
                        ...linesState.map(
                          (e) => Positioned(
                            left: 0,
                            top: 0,
                            child: LineMapItem(e),
                          ),
                        ),
                        ...nodesState.map(
                          (e) => Positioned(
                            left: e.position?.x ?? 0,
                            top: e.position?.y ?? 0,
                            child: NodeMapItem(e),
                          ),
                        ),
                        BlocBuilder<ConnectingLineCubit, LineModel?>(
                          builder: (context, state) {
                            if (state == null) {
                              return Container();
                            }
                            return Positioned(
                              left: 0,
                              top: 0,
                              child: LineMapItem(state),
                            );
                          },
                        ),
                        Positioned.fill(
                          child: DragTarget<NodeModel>(
                            builder: (context, candidateData, rejectedData) {
                              return Container(
                                color: Colors.transparent,
                              );
                            },
                            onWillAcceptWithDetails: (details) => true,
                            onAcceptWithDetails: (details) {
                              final renderBox = _canvasKey.currentContext!
                                  .findRenderObject() as RenderBox;
                              final localOffset =
                                  renderBox.globalToLocal(details.offset);

                              final node = details.data.copyWith(
                                position: NodePosition(
                                  x: localOffset.dx,
                                  y: localOffset.dy,
                                ),
                              );

                              BlocProvider.of<NodesCubit>(context)
                                  .addNode(node);
                            },
                          ),
                        ),
                      ],
                    );
                  },
                );
              },
            ),
          ),
        );
      },
    );
  }
}
