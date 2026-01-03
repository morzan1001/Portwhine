import 'dart:convert';

import 'package:portwhine/api/api.dart' as gen;
import 'package:portwhine/models/node_definition.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/pipeline_model.dart';

class SinglePipelineRepo {
  static Future<PipelineModel> getPipeline(String id) async {
    final response = await gen.api.apiV1PipelinePipelineIdGet(pipelineId: id);
    if (!response.isSuccessful) {
      throw Exception('Failed to load pipeline: ${response.statusCode}');
    }
    // PipelineOutput is empty (additionalProperties: true), so we parse the raw body
    final json = jsonDecode(response.bodyString) as Map<String, dynamic>;
    return PipelineModel.fromMap(json);
  }

  static Future<void> updatePipeline(PipelineModel pipeline) async {
    final response = await gen.api.apiV1PipelinePatch(
      body: pipeline.toPipelinePatch(),
    );
    if (!response.isSuccessful) {
      throw Exception('Failed to update pipeline: ${response.statusCode}');
    }
  }

  /// Get all available node definitions (triggers and workers).
  static Future<List<NodeDefinition>> getAllNodes() async {
    final response = await gen.api.apiV1NodesGet();
    if (!response.isSuccessful || response.body == null) {
      throw Exception('Failed to load nodes: ${response.statusCode}');
    }
    return response.body!.map(NodeDefinition.fromGenerated).toList();
  }

  /// Get all trigger node definitions.
  static Future<List<NodeDefinition>> getTriggerNodes() async {
    final response = await gen.api.apiV1NodesTriggersGet();
    if (!response.isSuccessful || response.body == null) {
      throw Exception('Failed to load triggers: ${response.statusCode}');
    }
    return response.body!.map(NodeDefinition.fromGenerated).toList();
  }

  /// Get all triggers as NodeModel list (for TriggersListBloc).
  static Future<List<NodeModel>> getAllTriggers() async {
    final definitions = await getTriggerNodes();
    return definitions.map(NodeModel.fromNodeDefinition).toList();
  }

  /// Get all worker node definitions.
  static Future<List<NodeDefinition>> getWorkerNodes() async {
    final response = await gen.api.apiV1NodesWorkersGet();
    if (!response.isSuccessful || response.body == null) {
      throw Exception('Failed to load workers: ${response.statusCode}');
    }
    return response.body!.map(NodeDefinition.fromGenerated).toList();
  }

  /// Get all workers as NodeModel list (for WorkersListBloc).
  static Future<List<NodeModel>> getAllWorkers() async {
    final definitions = await getWorkerNodes();
    return definitions.map(NodeModel.fromNodeDefinition).toList();
  }
}
