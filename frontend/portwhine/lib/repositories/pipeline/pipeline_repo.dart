import 'dart:math';

import 'package:portwhine/models/pipeline_model.dart';

class PipelineRepo {
  Future<int> getPipelineInProgress() async {
    await Future.delayed(const Duration(milliseconds: 2000));
    return 4;
  }

  Future<Map<DateTime, int>> getPipelineNumber() async {
    await Future.delayed(const Duration(milliseconds: 2000));
    return {
      DateTime(2023, 1): 20,
      DateTime(2023, 2): 15,
      DateTime(2023, 3): 12,
      DateTime(2023, 4): 17,
      DateTime(2023, 5): 5,
      DateTime(2023, 6): 8,
      DateTime(2023, 7): 12,
      DateTime(2023, 8): 15,
    };
  }

  Future<List<String>> getPipelineErrors() async {
    await Future.delayed(const Duration(milliseconds: 2000));
    return [
      'Nmap returned "NULL"',
      'Nmap returned "NULL"',
    ];
  }

  Future<List<PipelineModel>> getPipelineList() async {
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
        completed: Random().nextBool(),
        expectedTime: 3200,
        currentNode: 'Nmap',
        currentRunningTime: 1200,
        nodes: [],
      ),
    );
  }
}
