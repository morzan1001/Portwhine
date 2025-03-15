import 'package:flutter/material.dart';
import 'package:portwhine/pages/pipeline_details/canvas/triggers_list.dart';
import 'package:portwhine/pages/pipeline_details/canvas/workers_list.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodesList extends StatelessWidget {
  const NodesList({super.key});

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: [
        Expanded(child: WorkersList()),
        VerticalSpacer(24),
        Expanded(child: TriggersList()),
      ],
    );
  }
}
