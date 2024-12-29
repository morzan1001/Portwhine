import 'dart:math';

import 'package:portwhine/models/pipeline_model.dart';

class PipelinesRepo {
  Future<List<PipelineModel>> getPipelinesList() async {
    await Future.delayed(const Duration(milliseconds: 100));
    return List.generate(
      6,
      (index) => PipelineModel(
        name: 'Security Scan ${Random().nextInt(100)}',
        runningTime: 1655,
        status: 'Waiting for Nmap',
        totalNodes: Random().nextInt(4),
        nodesCompleted: Random().nextInt(4),
        errors: Random().nextInt(4),
        completed: false,
        expectedTime: 3200,
        currentNode: 'Nmap',
        currentRunningTime: 1200,
        nodes: [],
      ),
    );
  }
}
