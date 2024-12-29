import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/node_map_item.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/nodes_selection_list.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/pipeline_canvas.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/pipeline_controls.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/pipeline_zoom_controls.dart';

class PipelineCanvasWrapper extends StatefulWidget {
  const PipelineCanvasWrapper({super.key});

  @override
  State<PipelineCanvasWrapper> createState() => _PipelineCanvasWrapperState();
}

class _PipelineCanvasWrapperState extends State<PipelineCanvasWrapper>
    with SingleTickerProviderStateMixin {
  @override
  void initState() {
    Future.delayed(
      const Duration(milliseconds: 1),
      () {
        BlocProvider.of<CanvasCubit>(context).setController(this);
      },
    );
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<SelectedNodeCubit, NodeModel?>(
      builder: (context, selectedNode) {
        final canvas = BlocProvider.of<CanvasCubit>(context).state;

        return Stack(
          children: [
            // color filtered to blur canvas
            ColorFiltered(
              colorFilter: ColorFilter.mode(
                selectedNode == null ? Colors.transparent : Colors.black54,
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
                    child: Container(color: CustomColors.greyLighter),
                  ),

                  // main
                  const PipelineCanvas(),

                  // nodes list to drag on canvas
                  const Positioned(
                    top: 120,
                    left: 24,
                    child: NodesList(),
                  ),
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
            // Positioned(
            //     left: (selectedNode.position!.x + canvas.position.x) *
            //         canvas.zoom,
            //     top: (selectedNode.position!.y + canvas.position.y) *
            //         canvas.zoom,
            //     child: Transform.scale(
            //       scale: canvas.zoom,
            //       child: NodeMapItem(selectedNode),
            //     ),
            //   ),
          ],
        );
      },
    );
  }
}
