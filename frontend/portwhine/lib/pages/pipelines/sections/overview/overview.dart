import 'package:flutter/material.dart';
import 'package:frontend/pages/workflows/sections/overview/number_of_workflows.dart';
import 'package:frontend/pages/workflows/sections/overview/workflows_errors.dart';
import 'package:frontend/pages/workflows/sections/overview/workflows_in_progress.dart';
import 'package:frontend/widgets/spacer.dart';

class WorkflowsOverview extends StatelessWidget {
  const WorkflowsOverview({super.key});

  @override
  Widget build(BuildContext context) {
    return const IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          WorkflowsInProgress(),
          HorizontalSpacer(12),
          NumberOfWorkflows(),
          HorizontalSpacer(12),
          WorkflowsErrors(),
        ],
      ),
    );
  }
}
