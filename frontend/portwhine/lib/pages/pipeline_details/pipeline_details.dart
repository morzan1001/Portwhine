import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/pages/pipeline_details/sections/canvas/pipeline_canvas_wrapper.dart';
import 'package:portwhine/pages/pipeline_details/sections/status/pipeline_status.dart';

@RoutePage()
class PipelineDetailsPage extends StatefulWidget {
  const PipelineDetailsPage({
    @PathParam('id') required this.id,
    this.model,
    super.key,
  });

  final String id;
  final PipelineModel? model;

  @override
  State<PipelineDetailsPage> createState() => _PipelineDetailsPageState();
}

class _PipelineDetailsPageState extends State<PipelineDetailsPage> {
  @override
  void initState() {
    Future.delayed(
      const Duration(milliseconds: 0),
      () {
        if (widget.model != null) {
          BlocProvider.of<PipelineCubit>(context).setPipeline(
            widget.model!,
          );
        }
      },
    );
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Scaffold(
        endDrawer: const PipelineDetails(),
        body: BlocBuilder<PipelineCubit, PipelineModel>(
          builder: (context, state) {
            return const Column(
              children: [
                Expanded(child: PipelineCanvasWrapper()),
              ],
            );
          },
        ),
      ),
    );
  }
}
