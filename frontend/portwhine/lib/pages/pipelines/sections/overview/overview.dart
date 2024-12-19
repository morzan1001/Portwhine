import 'package:flutter/material.dart';
import 'package:frontend/pages/pipelines/sections/overview/number_of_pipelines.dart';
import 'package:frontend/pages/pipelines/sections/overview/pipeline_errors.dart';
import 'package:frontend/pages/pipelines/sections/overview/pipeline_in_progress.dart';
import 'package:frontend/widgets/spacer.dart';

class PipelinesOverview extends StatelessWidget {
  const PipelinesOverview({super.key});

  @override
  Widget build(BuildContext context) {
    return const IntrinsicHeight(
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
    );
  }
}
