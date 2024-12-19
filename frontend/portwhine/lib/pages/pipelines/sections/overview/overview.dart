import 'package:flutter/material.dart';
import 'package:frontend/pages/pipelines/sections/overview/number_of_pipelines.dart';
import 'package:frontend/pages/pipelines/sections/overview/pipeline_errors.dart';
import 'package:frontend/pages/pipelines/sections/overview/pipeline_in_progress.dart';
import 'package:frontend/widgets/spacer.dart';
import 'package:frontend/global/colors.dart';
import 'package:frontend/widgets/button.dart';

class PipelinesOverview extends StatelessWidget {
  const PipelinesOverview({super.key});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Button(
              'Create Pipeline',
              height: 44,
              icon: Icons.add,
              onPressed: () {},
            ),
            const HorizontalSpacer(12),
            Button(
              'Delete Pipeline',
              height: 44,
              icon: Icons.delete,
              buttonColor: CustomColors.red,
              onPressed: () {},
            ),
          ],
        ),
        const VerticalSpacer(12),
        const IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              PipelinesInProgress(),
              HorizontalSpacer(12),
              NumberOfPipelines(),
              HorizontalSpacer(12),
              PipelinesErrors(),
            ],
          ),
        ),
      ],
    );
  }
}
