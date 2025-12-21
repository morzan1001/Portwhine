import 'package:portwhine/api/api.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/pipeline_model.dart';

class SinglePipelineRepo {
  static Future<PipelineModel> getPipeline(String id) async {
    final result = await Api.getPipeline(id);
    return result;
  }

  static Future<void> updatePipeline(PipelineModel pipeline) async {
    await Api.updatePipeline(pipeline.toMap());
  }

  static Future<List<NodeModel>> getAllWorkers() async {
    final names = await Api.getAllWorkers();
    final List<NodeModel> workers = [];

    for (final name in names) {
      final config = await Api.getWorkerConfig(name);
      final example = config['example'] as Map<String, dynamic>;

      // Parse inputs/outputs from example
      // example['input'] is List<String> e.g. ["ip"]
      final inputsList = (example['input'] as List?)?.cast<String>() ?? [];
      final outputsList = (example['output'] as List?)?.cast<String>() ?? [];

      final inputs = {for (var i in inputsList) i: String};
      final outputs = {for (var i in outputsList) i: String};

      workers.add(NodeModel(
        name: name,
        inputs: inputs,
        outputs: outputs,
        config: example, // Store the full example as initial config
      ));
    }
    return workers;
  }

  static Future<List<NodeModel>> getAllTriggers() async {
    final names = await Api.getAllTriggers();
    final List<NodeModel> triggers = [];

    for (final name in names) {
      final config = await Api.getTriggerConfig(name);
      final example = config['example'] as Map<String, dynamic>;

      // Triggers usually have outputs but no inputs (except maybe from other triggers?)
      final outputsList = (example['output'] as List?)?.cast<String>() ?? [];
      final outputs = {for (var i in outputsList) i: String};

      triggers.add(NodeModel(
        name: name,
        inputs: {},
        outputs: outputs,
        config: example,
      ));
    }
    return triggers;
  }
}
