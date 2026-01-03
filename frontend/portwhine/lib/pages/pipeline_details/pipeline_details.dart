import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/nodes/node_definitions_cubit.dart';
import 'package:portwhine/blocs/pipeline_status/pipeline_status_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/show_nodes/show_nodes_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/single_pipeline/single_pipeline_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/triggers_list/triggers_list_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/workers_list/workers_list_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/pages/pipeline_details/canvas/node_map_item.dart';
import 'package:portwhine/pages/pipeline_details/canvas/nodes_list.dart';
import 'package:portwhine/pages/pipeline_details/canvas/pipeline_canvas.dart';
import 'package:portwhine/pages/pipeline_details/canvas/pipeline_controls.dart';
import 'package:portwhine/pages/pipeline_details/canvas/pipeline_zoom_controls.dart';
import 'package:portwhine/pages/pipeline_details/widgets/pipeline_info_drawer.dart';
import 'package:portwhine/pages/pipeline_details/widgets/shadow_container.dart';
import 'package:portwhine/utils/pipeline_error_parser.dart';
import 'package:portwhine/widgets/toast/toast_service.dart';

import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';

@RoutePage()
class PipelineDetailsPage extends StatefulWidget {
  const PipelineDetailsPage({@PathParam('id') required this.id, super.key});

  final String id;

  @override
  State<PipelineDetailsPage> createState() => _PipelineDetailsPageState();
}

class _PipelineDetailsPageState extends State<PipelineDetailsPage>
    with SingleTickerProviderStateMixin {
  PipelineStatusCubit? _pipelineStatusCubit;

  @override
  void initState() {
    super.initState();

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<CanvasCubit>().setController(this);
      context.read<SinglePipelineBloc>().add(GetSinglePipeline(widget.id));
      context.read<WorkersListBloc>().add(GetWorkersList());
      context.read<TriggersListBloc>().add(GetTriggersList());

      // Connect to WebSocket for live status updates
      _pipelineStatusCubit?.connectToPipeline(widget.id);
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Cache cubit reference for safe use in dispose (no ancestor lookup then).
    _pipelineStatusCubit ??= context.read<PipelineStatusCubit>();
  }

  @override
  void dispose() {
    // Disconnect WebSocket when leaving the page
    _pipelineStatusCubit?.disconnect();
    super.dispose();
  }

  /// Enrich nodes with definitions and live status info
  List<NodeModel> _enrichNodes(
    List<NodeModel> nodes,
    NodeDefinitionsCubit definitionsCubit,
    PipelineStatusCubit statusCubit,
  ) {
    return nodes.map((node) {
      // Get definition for this node type
      final definition = definitionsCubit.getNodeById(node.name);

      // Get live status info
      final statusInfo = statusCubit.getNodeStatus(node.id);

      return node.copyWith(definition: definition, statusInfo: statusInfo);
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Scaffold(
        endDrawer: const PipelineInfoDrawer(),
        body: MultiBlocListener(
          listeners: [
            BlocListener<SinglePipelineBloc, SinglePipelineState>(
              listener: (context, state) {
                if (state is SinglePipelineFailed) {
                  ToastService.error(context, state.error);

                  // Parse errors and update nodes
                  final nodesCubit = context.read<NodesCubit>();
                  final currentNodes = nodesCubit.state;
                  final nodeErrors = PipelineErrorParser.parseErrors(
                    state.error,
                    currentNodes,
                  );

                  if (nodeErrors.isNotEmpty) {
                    final updatedNodes = currentNodes.map((node) {
                      if (nodeErrors.containsKey(node.id)) {
                        return node.copyWith(error: nodeErrors[node.id]);
                      }
                      // Keep existing error or clear it?
                      // For now, let's assume a new save attempt clears old errors
                      // unless they are re-detected.
                      return node.copyWith(error: '');
                    }).toList();
                    nodesCubit.setNodes(updatedNodes);
                  }
                }
                if (state is SinglePipelineLoaded) {
                  // Enrich nodes with definitions
                  final definitionsCubit = context.read<NodeDefinitionsCubit>();
                  final statusCubit = context.read<PipelineStatusCubit>();
                  final enrichedNodes = _enrichNodes(
                    state.pipeline.nodes,
                    definitionsCubit,
                    statusCubit,
                  );

                  // Clear any previous errors on successful load/save
                  final cleanNodes = enrichedNodes
                      .map((n) => n.copyWith(error: ''))
                      .toList();

                  final enrichedPipeline = state.pipeline.copyWith(
                    nodes: cleanNodes,
                  );

                  context.read<PipelineCubit>().setPipeline(enrichedPipeline);
                  context.read<NodesCubit>().setNodes(cleanNodes);
                  context.read<LinesCubit>().updateLines(cleanNodes);
                }
              },
            ),
            // Listen for live status updates and refresh nodes
            BlocListener<PipelineStatusCubit, PipelineStatusState>(
              listener: (context, state) {
                if (state is PipelineStatusConnected) {
                  // Update node status info when we receive WebSocket updates
                  final currentNodes = context.read<NodesCubit>().state;
                  final definitionsCubit = context.read<NodeDefinitionsCubit>();
                  final statusCubit = context.read<PipelineStatusCubit>();
                  final enrichedNodes = _enrichNodes(
                    currentNodes,
                    definitionsCubit,
                    statusCubit,
                  );
                  context.read<NodesCubit>().setNodes(enrichedNodes);
                }
              },
            ),
          ],
          child: BlocBuilder<SelectedNodeCubit, NodeModel?>(
            builder: (context, selectedNode) {
              final canvas = BlocProvider.of<CanvasCubit>(context).state;

              return Stack(
                children: [
                  // color filtered to blur canvas
                  ColorFiltered(
                    colorFilter: ColorFilter.mode(
                      selectedNode == null
                          ? Colors.transparent
                          : Colors.black54,
                      BlendMode.multiply,
                    ),
                    child: Stack(
                      children: [
                        // bg container to show grey color
                        Positioned(
                          top: 0,
                          left: 0,
                          right: 0,
                          bottom: 0,
                          child: Container(color: MyColors.lightGrey),
                        ),

                        // main
                        const PipelineCanvas(),

                        // nodes list to drag on canvas
                        BlocBuilder<ShowNodesCubit, bool>(
                          builder: (context, state) {
                            return AnimatedPositioned(
                              top: 104,
                              left: state ? 24 : -304,
                              bottom: 104,
                              duration: const Duration(milliseconds: 100),
                              child: const NodesList(),
                            );
                          },
                        ),

                        // nodes list controller
                        BlocBuilder<ShowNodesCubit, bool>(
                          builder: (context, state) {
                            return Positioned(
                              bottom: 24,
                              left: 24,
                              child: ShadowContainer(
                                onTap: () {
                                  BlocProvider.of<ShowNodesCubit>(
                                    context,
                                  ).toggleNodes();
                                },
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 18,
                                ),
                                child: Icon(
                                  state
                                      ? Icons.chevron_left
                                      : Icons.chevron_right,
                                  color: MyColors.black,
                                  size: 20,
                                ),
                              ),
                            );
                          },
                        ),

                        // controls
                        const Positioned(
                          top: 24,
                          left: 24,
                          right: 24,
                          child: PipelineControls(),
                        ),
                        const Positioned(
                          bottom: 24,
                          right: 24,
                          child: PipelineZoomControls(),
                        ),
                      ],
                    ),
                  ),

                  // showing selected node when dialog is open
                  if (selectedNode != null)
                    Positioned(
                      left: (width(context) - nodeWidth) / 2,
                      top: 100,
                      child: Transform.scale(
                        scale: canvas.zoom,
                        child: NodeMapItem(selectedNode),
                      ),
                    ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}
