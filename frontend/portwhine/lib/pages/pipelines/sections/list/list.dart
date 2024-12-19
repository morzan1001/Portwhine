import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:frontend/blocs/workflows/workflows_list/workflows_list_bloc.dart';
import 'package:frontend/global/colors.dart';
import 'package:frontend/global/text_style.dart';
import 'package:frontend/pages/workflows/sections/list/workflow_item.dart';
import 'package:frontend/widgets/button.dart';
import 'package:frontend/widgets/loading_indicator.dart';
import 'package:frontend/widgets/spacer.dart';
import 'package:shimmer/shimmer.dart';

class PipelinesList extends StatelessWidget {
  const PipelinesList({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: 24,
        vertical: 16,
      ),
      decoration: BoxDecoration(
        color: CustomColors.greyVar,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // heading and buttons
          Row(
            children: [
              Expanded(
                child: Text(
                  'Pipelines',
                  style: style(
                    size: 24,
                    weight: FontWeight.w600,
                    color: CustomColors.textDark,
                  ),
                ),
              ),
              Button(
                'Export to CSV',
                height: 44,
                icon: Icons.folder,
                buttonColor: CustomColors.white,
                borderColor: CustomColors.grey,
                onPressed: () {},
              ),
              const HorizontalSpacer(12),
              Button(
                'Filter',
                height: 44,
                icon: Icons.tune,
                buttonColor: CustomColors.white,
                borderColor: CustomColors.grey,
                onPressed: () {},
              ),
              const HorizontalSpacer(12),
              Button(
                'Add new',
                height: 44,
                icon: Icons.add,
                onPressed: () {},
              ),
            ],
          ),
          const VerticalSpacer(24),

          BlocBuilder<WorkflowsListBloc, WorkflowsListState>(
            builder: (context, state) {
              if (state is WorkflowsListLoading) {
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

              if (state is WorkflowsListFailed) {
                return Text(state.error);
              }

              if (state is WorkflowsListLoaded) {
                final completedWorkflows =
                    state.workflows.where((e) => e.completed).toList();
                final otherWorkflows =
                    state.workflows.where((e) => !e.completed).toList();

                return Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // completed workflows
                    Text(
                      'Completed',
                      style: style(
                        color: CustomColors.textLight,
                        weight: FontWeight.w400,
                        size: 13,
                      ),
                    ),
                    const VerticalSpacer(12),
                    ListView.separated(
                      separatorBuilder: (a, b) => const VerticalSpacer(12),
                      itemCount: completedWorkflows.length,
                      shrinkWrap: true,
                      itemBuilder: (_, i) {
                        return WorkflowItem(completedWorkflows[i]);
                      },
                    ),
                    const VerticalSpacer(24),

                    // other workflows
                    Text(
                      'Others',
                      style: style(
                        color: CustomColors.textLight,
                        weight: FontWeight.w400,
                        size: 13,
                      ),
                    ),
                    const VerticalSpacer(12),
                    ListView.separated(
                      separatorBuilder: (a, b) => const VerticalSpacer(12),
                      itemCount: otherWorkflows.length,
                      shrinkWrap: true,
                      itemBuilder: (_, i) {
                        return WorkflowItem(otherWorkflows[i]);
                      },
                    ),
                  ],
                );
              }

              return const LoadingIndicator();
            },
          ),
        ],
      ),
    );
  }
}
