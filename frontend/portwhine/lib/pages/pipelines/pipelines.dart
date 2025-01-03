import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/bloc_listeners.dart';
import 'package:portwhine/blocs/pipelines/get_all_pipelines/get_all_pipelines_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/pages/pipelines/change_page_section.dart';
import 'package:portwhine/pages/pipelines/pipeline_item.dart';
import 'package:portwhine/pages/write_pipeline/write_pipeline.dart';
import 'package:portwhine/widgets/button.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/svg_icon.dart';
import 'package:portwhine/widgets/text.dart';
import 'package:shimmer/shimmer.dart';

@RoutePage()
class PipelinesPage extends StatelessWidget {
  const PipelinesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: BlocListeners.pipelinesListener,
      child: SafeArea(
        child: Scaffold(
          backgroundColor: MyColors.lightGrey,
          body: Container(
            padding: const EdgeInsets.symmetric(
              horizontal: 24,
              vertical: 24,
            ),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // logo
                const SvgIcon(icon: 'logo', size: 72),
                const VerticalSpacer(24),

                // heading and add button
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Heading('Pipelines', size: 24, bold: true),
                    Button(
                      'Add new',
                      height: 44,
                      icon: Icons.add,
                      onTap: () => showWritePipelineDialog(context),
                    ),
                  ],
                ),
                const VerticalSpacer(24),

                // list of pipelines
                Expanded(
                  child: BlocBuilder<GetAllPipelinesBloc, GetAllPipelinesState>(
                    builder: (context, state) {
                      if (state is GetAllPipelinesLoading) {
                        return Shimmer.fromColors(
                          baseColor: MyColors.grey,
                          highlightColor: MyColors.darkGrey,
                          child: Column(
                            children: [
                              ...List.generate(
                                3,
                                (i) => Container(
                                  margin: const EdgeInsets.only(bottom: 12),
                                  height: 76,
                                  width: double.infinity,
                                  decoration: BoxDecoration(
                                    borderRadius: BorderRadius.circular(12),
                                    color: MyColors.white,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        );
                      }

                      if (state is GetAllPipelinesFailed) {
                        return Text(state.error);
                      }

                      if (state is GetAllPipelinesLoaded) {
                        final pipelines = state.pipelines;

                        return ListView.separated(
                          separatorBuilder: (a, b) => const VerticalSpacer(12),
                          itemCount: pipelines.length,
                          shrinkWrap: true,
                          itemBuilder: (_, i) {
                            return PipelineItem(pipelines[i]);
                          },
                        );
                      }

                      return const SizedBox.shrink();
                    },
                  ),
                ),
                const VerticalSpacer(24),

                // change page
                const ChangePageSection(),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
