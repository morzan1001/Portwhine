import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/single_pipeline/single_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/pages/pipeline_details/results_page.dart';
import 'package:portwhine/pages/pipeline_details/widgets/shadow_container.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/toast.dart';
import 'package:url_launcher/url_launcher.dart';

class PipelineControls extends StatelessWidget {
  const PipelineControls({super.key});

  Future<void> _launchKibana(BuildContext context, String pipelineId) async {
    // Kibana URL with a filter for the current pipeline ID
    // We assume Kibana is running on localhost:5601 and using the default index pattern or one that covers pipeline_results
    // The query uses KQL (Kibana Query Language)
    final kibanaUrl = Uri.parse(
      'https://kibana.portwhine.local/app/discover#/?_a=(query:(language:kuery,query:\'pipeline_id:"$pipelineId"\'))',
    );

    try {
      if (!await launchUrl(kibanaUrl, mode: LaunchMode.externalApplication)) {
        if (context.mounted) {
          showToast(context, 'Could not launch Kibana');
        }
      }
    } catch (e) {
      if (context.mounted) {
        showToast(context, 'Error launching Kibana: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineCubit, PipelineModel>(
      builder: (context, state) {
        final isRunning = state.status == kStatusRunning;

        return Row(
          children: [
            // back button
            ShadowContainer(
              onTap: () => pop(context),
              padding: const EdgeInsets.symmetric(horizontal: 18),
              child: const Icon(
                Icons.arrow_back,
                color: MyColors.black,
                size: 20,
              ),
            ),

            const Spacer(),

            // buttons
            ShadowContainer(
              onTap: () => _launchKibana(context, state.id),
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Center(
                child: Text(
                  'View Results',
                  style: style(
                    color: MyColors.textDarkGrey,
                    weight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const HorizontalSpacer(16),
            ShadowContainer(
              onTap: () {
                final pipeline = context.read<PipelineCubit>().state;
                final nodes = context.read<NodesCubit>().state;

                final updatedPipeline = pipeline.copyWith(
                  nodes: nodes,
                );

                context
                    .read<SinglePipelineBloc>()
                    .add(UpdatePipeline(updatedPipeline));
              },
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Center(
                child: Text(
                  'Save',
                  style: style(
                    color: MyColors.textDarkGrey,
                    weight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const HorizontalSpacer(16),

            // controls
            BlocBuilder<CanvasCubit, CanvasModel>(
              builder: (context, canvasState) {
                return ShadowContainer(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    children: [
                      InkWell(
                        onTap: isRunning
                            ? () {
                                context
                                    .read<StartStopPipelineBloc>()
                                    .add(StopPipeline(state.id));
                              }
                            : null,
                        child: Icon(
                          Icons.pause_outlined,
                          color: isRunning ? MyColors.red : MyColors.grey,
                          size: 20,
                        ),
                      ),
                      const HorizontalSpacer(20),
                      InkWell(
                        onTap: !isRunning
                            ? () {
                                context
                                    .read<StartStopPipelineBloc>()
                                    .add(StartPipeline(state.id));
                              }
                            : null,
                        child: Icon(
                          Icons.play_arrow_outlined,
                          color: !isRunning ? MyColors.green : MyColors.grey,
                          size: 20,
                        ),
                      ),
                      const HorizontalSpacer(20),
                      Container(
                        height: 32,
                        width: 1,
                        color: MyColors.darkGrey,
                      ),
                      const HorizontalSpacer(20),
                      InkWell(
                        child: const Icon(
                          Icons.info_outline,
                          color: MyColors.black,
                          size: 20,
                        ),
                        onTap: () {
                          Scaffold.of(context).openEndDrawer();
                        },
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        );
      },
    );
  }
}
