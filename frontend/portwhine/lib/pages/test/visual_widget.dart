import 'package:flutter/material.dart';

class PipelineVisualization extends StatefulWidget {
  final Map<String, dynamic> pipelineData;

  const PipelineVisualization({super.key, required this.pipelineData});

  @override
  State<PipelineVisualization> createState() => _PipelineVisualizationState();
}

class _PipelineVisualizationState extends State<PipelineVisualization> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: SingleChildScrollView(
          child: Container(
            padding: const EdgeInsets.all(20),
            child: IntrinsicHeight(
              child: Row(
                mainAxisAlignment: MainAxisAlignment.start,
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  _buildTrigger(widget.pipelineData['trigger']),
                  const SizedBox(width: 50),
                  _buildWorkerColumn(widget.pipelineData['worker']),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTrigger(Map<String, dynamic> triggerData) {
    String triggerType = triggerData.keys.first;
    var trigger = triggerData[triggerType];

    return PipelineNode(
      title: triggerType,
      inputs: const [],
      outputs: List<String>.from(trigger['output']),
    );
  }

  Widget _buildWorkerColumn(List<dynamic> workers) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: workers.map((worker) {
        return Padding(
          padding: EdgeInsets.symmetric(
            vertical: workers.length > 1 ? 20.0 : 0,
          ),
          child: _buildWorkerRow(worker),
        );
      }).toList(),
    );
  }

  Widget _buildWorkerRow(dynamic worker) {
    var workerData = worker.values.first;
    var workerType = worker.keys.first.toString();
    var children = workerData['children'] as List?;

    return IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          PipelineNode(
            title: workerType,
            inputs: List<String>.from(workerData['input']),
            outputs: List<String>.from(workerData['output']),
          ),
          if (children != null && children.isNotEmpty) ...[
            const SizedBox(width: 50),
            _buildChildrenColumn(children),
          ],
        ],
      ),
    );
  }

  Widget _buildChildrenColumn(List<dynamic> children) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: children.map((child) {
        return Padding(
          padding: EdgeInsets.symmetric(
            vertical: children.length > 1 ? 20.0 : 0,
          ),
          child: _buildWorkerRow(child),
        );
      }).toList(),
    );
  }
}

class PipelineNode extends StatelessWidget {
  final String title;
  final List<String> inputs;
  final List<String> outputs;

  const PipelineNode({
    super.key,
    required this.title,
    required this.inputs,
    required this.outputs,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(
        minWidth: 150,
        minHeight: 80,
      ),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        border: Border.all(color: Colors.black),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 14,
            ),
            textAlign: TextAlign.center,
          ),
          if (inputs.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text(
              'Input: ${inputs.join(", ")}',
              style: const TextStyle(fontSize: 12),
              textAlign: TextAlign.center,
            ),
          ],
          if (outputs.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              'Output: ${outputs.join(", ")}',
              style: const TextStyle(fontSize: 12),
              textAlign: TextAlign.center,
            ),
          ],
        ],
      ),
    );
  }
}
