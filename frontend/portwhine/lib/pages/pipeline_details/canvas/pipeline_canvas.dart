import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/colors.dart';
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
  final GlobalKey _interactiveViewerKey = GlobalKey();
  final GlobalKey _canvasKey = GlobalKey();

  // Cache the grid painter to avoid recreation
  static const _gridPainter = GridPainter();

  @override
  void initState() {
    super.initState();
    // Register the interactive viewer key after the first frame
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<CanvasCubit>().setCanvasKey(_interactiveViewerKey);
    });
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CanvasCubit, CanvasModel>(
      builder: (context, state) {
        final controller = state.controller;

        return InteractiveViewer(
          key: _interactiveViewerKey,
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
                      clipBehavior: Clip.none,
                      children: [
                        // Grid background - static, cached painter
                        Positioned.fill(
                          child: RepaintBoundary(
                            child: Container(
                              decoration: BoxDecoration(
                                gradient: RadialGradient(
                                  center: Alignment.center,
                                  radius: 2,
                                  colors: [
                                    MyColors.lightGrey,
                                    MyColors.lightGrey.withValues(alpha: 0.95),
                                  ],
                                ),
                              ),
                              child: CustomPaint(
                                painter: _gridPainter,
                                isComplex: true,
                                willChange: false,
                              ),
                            ),
                          ),
                        ),
                        // Connection lines
                        ...linesState.map(
                          (e) => Positioned(
                            left: 0,
                            top: 0,
                            child: LineMapItem(e),
                          ),
                        ),
                        // Drop target for new nodes - below nodes in stack order
                        Positioned.fill(
                          child: DragTarget<NodeModel>(
                            builder: (context, candidateData, rejectedData) {
                              // This container is just for visual feedback, doesn't block events
                              return Container(
                                color: candidateData.isNotEmpty
                                    ? const Color(0xFF6366F1)
                                        .withValues(alpha: 0.05)
                                    : Colors.transparent,
                              );
                            },
                            onWillAcceptWithDetails: (details) => true,
                            onAcceptWithDetails: (details) {
                              final canvas = context.read<CanvasCubit>().state;
                              final localOffset =
                                  canvas.globalToCanvas(details.offset);

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
                        // Nodes - use keys for efficient updates
                        ...nodesState.map(
                          (e) => Positioned(
                            key: ValueKey(e.id),
                            left: e.position?.x ?? 0,
                            top: e.position?.y ?? 0,
                            child: NodeMapItem(e),
                          ),
                        ),
                        // Active connecting line
                        BlocBuilder<ConnectingLineCubit, LineModel?>(
                          builder: (context, lineState) {
                            if (lineState == null) {
                              return const SizedBox.shrink();
                            }
                            return Positioned.fill(
                              child: IgnorePointer(
                                child:
                                    LineMapItem(lineState, isConnecting: true),
                              ),
                            );
                          },
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
