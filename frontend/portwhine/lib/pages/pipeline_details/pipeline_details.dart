import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
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
import 'package:portwhine/pages/pipeline_details/widgets/shadow_container.dart';

import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';

@RoutePage()
class PipelineDetailsPage extends StatefulWidget {
  const PipelineDetailsPage({
    @PathParam('id') required this.id,
    super.key,
  });

  final String id;

  @override
  State<PipelineDetailsPage> createState() => _PipelineDetailsPageState();
}

class _PipelineDetailsPageState extends State<PipelineDetailsPage>
    with SingleTickerProviderStateMixin {
  @override
  void initState() {
    Future.delayed(
      const Duration(milliseconds: 0),
      () {
        context.read<CanvasCubit>().setController(this);
        context.read<SinglePipelineBloc>().add(GetSinglePipeline(widget.id));
        context.read<WorkersListBloc>().add(GetWorkersList());
        context.read<TriggersListBloc>().add(GetTriggersList());
      },
    );
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Scaffold(
        body: MultiBlocListener(
          listeners: [
            BlocListener<SinglePipelineBloc, SinglePipelineState>(
              listener: (context, state) {
                if (state is SinglePipelineLoaded) {
                  context.read<PipelineCubit>().setPipeline(state.pipeline);
                  context.read<NodesCubit>().setNodes(state.pipeline.nodes);
                  // LinesCubit needs to be updated based on nodes, but it calculates lines from nodes.
                  // We can trigger an update or set lines directly if we had setLines.
                  // But LinesCubit.updateLines takes nodes and recalculates.
                  context.read<LinesCubit>().updateLines(state.pipeline.nodes);
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
                                  BlocProvider.of<ShowNodesCubit>(context)
                                      .toggleNodes();
                                },
                                padding:
                                    const EdgeInsets.symmetric(horizontal: 18),
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
