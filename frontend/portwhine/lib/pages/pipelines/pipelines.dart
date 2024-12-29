import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/pages/pipelines/pipeline_item.dart';
import 'package:portwhine/widgets/button.dart';
import 'package:portwhine/widgets/loading_indicator.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/svg_icon.dart';
import 'package:shimmer/shimmer.dart';

@RoutePage()
class PipelinesPage extends StatelessWidget {
  const PipelinesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: CustomColors.greyVar,
      body: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: 24,
          vertical: 16,
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
                Text(
                  'Pipelines',
                  style: style(
                    size: 24,
                    weight: FontWeight.w600,
                    color: CustomColors.textDark,
                  ),
                ),
                Button(
                  'Add new',
                  height: 44,
                  icon: Icons.add,
                  onPressed: () {},
                ),
              ],
            ),
            const VerticalSpacer(24),

            // list of pipelines
            BlocBuilder<PipelinesListBloc, PipelinesListState>(
              builder: (context, state) {
                if (state is PipelinesListLoading) {
                  return Shimmer.fromColors(
                    baseColor: CustomColors.greyVar,
                    highlightColor: CustomColors.grey,
                    child: Column(
                      children: [
                        ...List.generate(
                          3,
                          (i) => Container(
                            margin: const EdgeInsets.symmetric(vertical: 6),
                            height: 40,
                            width: double.infinity,
                            decoration: BoxDecoration(
                              borderRadius: BorderRadius.circular(16),
                              color: CustomColors.white,
                            ),
                          ),
                        ),
                      ],
                    ),
                  );
                }

                if (state is PipelinesListFailed) {
                  return Text(state.error);
                }

                if (state is PipelinesListLoaded) {
                  final otherPipelines =
                      state.pipelines.where((e) => !e.completed).toList();

                  return ListView.separated(
                    separatorBuilder: (a, b) => const VerticalSpacer(12),
                    itemCount: otherPipelines.length,
                    shrinkWrap: true,
                    itemBuilder: (_, i) {
                      return PipelineItem(otherPipelines[i]);
                    },
                  );
                }

                return const LoadingIndicator();
              },
            ),
          ],
        ),
      ),
    );
  }
}
