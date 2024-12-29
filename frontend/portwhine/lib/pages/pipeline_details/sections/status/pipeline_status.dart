import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineDetails extends StatelessWidget {
  const PipelineDetails({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineCubit, PipelineModel>(
      builder: (context, state) {
        return Drawer(
          child: Container(
            padding: const EdgeInsets.all(24),
            color: CustomColors.white,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Pipeline Status',
                        style: style(
                          color: CustomColors.textDark,
                          weight: FontWeight.w600,
                          size: 20,
                        ),
                      ),
                    ),
                    InkWell(
                      onTap: () {
                        Scaffold.of(context).closeEndDrawer();
                      },
                      child: const Icon(Icons.close, size: 20),
                    ),
                  ],
                ),
                const VerticalSpacer(40),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Node Running',
                        style: style(
                          color: CustomColors.textLight,
                          weight: FontWeight.w400,
                        ),
                      ),
                    ),
                    Text(
                      state.currentNode ?? '',
                      style: style(
                        color: CustomColors.textDark,
                        weight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const VerticalSpacer(24),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Current Runtime',
                        style: style(
                          color: CustomColors.textLight,
                          weight: FontWeight.w400,
                        ),
                      ),
                    ),
                    Text(
                      '${state.runningTime}',
                      style: style(
                        color: CustomColors.textDark,
                        weight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const VerticalSpacer(24),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Errors Reported',
                        style: style(
                          color: CustomColors.textLight,
                          weight: FontWeight.w400,
                        ),
                      ),
                    ),
                    Text(
                      '${state.errors}',
                      style: style(
                        color: CustomColors.textDark,
                        weight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const VerticalSpacer(24),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Expected Finish Time',
                        style: style(
                          color: CustomColors.textLight,
                          weight: FontWeight.w400,
                        ),
                      ),
                    ),
                    Text(
                      '${state.expectedTime}',
                      style: style(
                        color: CustomColors.textDark,
                        weight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
