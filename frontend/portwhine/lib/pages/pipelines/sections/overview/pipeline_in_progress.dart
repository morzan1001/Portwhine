import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:frontend/blocs/workflows/workflows_in_progress/workflows_in_progress_bloc.dart';
import 'package:frontend/global/colors.dart';
import 'package:frontend/global/text_style.dart';
import 'package:frontend/widgets/loading_indicator.dart';
import 'package:frontend/widgets/spacer.dart';
import 'package:shimmer/shimmer.dart';

class PipelinesInProgress extends StatelessWidget {
  const PipelinesInProgress({super.key});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: 24,
          vertical: 16,
        ),
        decoration: BoxDecoration(
          color: CustomColors.white,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Pipelines in progress',
              style: style(
                size: 14,
                color: CustomColors.textDark,
                weight: FontWeight.w600,
              ),
            ),
            const VerticalSpacer(12),
            BlocBuilder<WorkflowsInProgressBloc, WorkflowsInProgressState>(
              builder: (context, state) {
                if (state is WorkflowsInProgressLoading) {
                  return Shimmer.fromColors(
                    baseColor: CustomColors.greyVar,
                    highlightColor: CustomColors.grey,
                    child: Container(
                      height: 36,
                      width: 36,
                      color: Colors.white,
                    ),
                  );
                }

                if (state is WorkflowsInProgressFailed) {
                  return Text(state.error);
                }

                if (state is WorkflowsInProgressLoaded) {
                  return Text(
                    state.number.toString(),
                    style: style(
                      size: 36,
                      color: CustomColors.sec,
                      weight: FontWeight.w600,
                    ),
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
