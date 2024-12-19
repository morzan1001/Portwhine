import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:frontend/blocs/workflows/workflows_errors/workflows_errors_bloc.dart';
import 'package:frontend/global/colors.dart';
import 'package:frontend/global/text_style.dart';
import 'package:frontend/widgets/loading_indicator.dart';
import 'package:frontend/widgets/spacer.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shimmer/shimmer.dart';

class WorkflowsErrors extends StatelessWidget {
  const WorkflowsErrors({super.key});

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
              'Errors and Warnings',
              style: style(
                size: 14,
                color: CustomColors.textDark,
                weight: FontWeight.w600,
              ),
            ),
            BlocBuilder<WorkflowsErrorsBloc, WorkflowsErrorsState>(
              builder: (context, state) {
                if (state is WorkflowsErrorsLoading) {
                  return Shimmer.fromColors(
                    baseColor: CustomColors.greyVar,
                    highlightColor: CustomColors.grey,
                    child: Padding(
                      padding: const EdgeInsets.only(top: 12),
                      child: Column(
                        children: [
                          ...List.generate(
                            3,
                            (i) => Container(
                              margin: const EdgeInsets.symmetric(vertical: 6),
                              height: 10,
                              width: double.infinity,
                              color: Colors.white,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }

                if (state is WorkflowsErrorsFailed) {
                  return Text(state.error);
                }

                if (state is WorkflowsErrorsLoaded) {
                  return Column(
                    children: List.generate(
                      state.errors.length,
                      (i) => ErrorItem(state.errors[i]),
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

class ErrorItem extends StatelessWidget {
  const ErrorItem(this.error, {super.key});

  final String error;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 12),
      child: Row(
        children: [
          Container(
            height: 10,
            width: 10,
            decoration: const BoxDecoration(
              color: CustomColors.error,
              shape: BoxShape.circle,
            ),
          ),
          const HorizontalSpacer(8),
          Container(
            padding: const EdgeInsets.symmetric(
              vertical: 4,
              horizontal: 6,
            ),
            decoration: BoxDecoration(
              color: CustomColors.greyLighter,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              error.split(' ')[0],
              style: GoogleFonts.spaceMono(
                fontSize: 13,
              ),
            ),
          ),
          const HorizontalSpacer(4),
          Flexible(
            child: Text(
              error.substring(error.split(' ')[0].length).trim(),
              overflow: TextOverflow.ellipsis,
              style: style(
                size: 13,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
