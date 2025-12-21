import 'dart:convert';

import 'package:portwhine/api/api_service.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/models/pipeline_model.dart';

class Api {
  static late ApiService service;

  static init() async => service = await ApiService.create();

  static Future<List<PipelineModel>> getAllPipelines({
    int size = 10,
    int page = 1,
  }) async {
    final result = await service.getAllPipelines(page: page, size: size);
    return (result.body ?? []).map(PipelineModel.fromMap).toList();
  }

  static Future<PipelineModel> createPipeline(String name) async {
    final result = await service.createPipeline(
      {'name': name, 'trigger': {}, 'worker': []},
    );
    return PipelineModel.fromMap(result.body!);
  }

  static Future<bool> deletePipeline(String id) async {
    final result = await service.deletePipeline(id);
    return result.body?['detail'] == kPipelineDeleted;
  }

  static Future<Map<String, dynamic>> startPipeline(String id) async {
    final result = await service.startPipeline(id);
    if (result.error != null) return jsonDecode(result.error! as String);
    return result.body ?? defaultErrorMap;
  }

  static Future<Map<String, dynamic>> stopPipeline(String id) async {
    final result = await service.stopPipeline(id);
    if (result.error != null) return jsonDecode(result.error! as String);
    return result.body ?? defaultErrorMap;
  }

  static Future<PipelineModel> getPipeline(String id) async {
    final result = await service.getPipeline(id);
    return PipelineModel.fromMap(result.body!);
  }

  static Future<PipelineModel> updatePipeline(
      Map<String, dynamic> pipeline) async {
    final result = await service.updatePipeline(pipeline);
    return PipelineModel.fromMap(result.body!);
  }

  static Future<List<String>> getAllWorkers() async {
    final result = await service.getAllWorkers();
    return result.body!;
  }

  static Future<Map<String, dynamic>> getWorkerConfig(String name) async {
    final result = await service.getWorkerConfig(name);
    return result.body!;
  }

  static Future<List<String>> getAllTriggers() async {
    final result = await service.getAllTriggers();
    return result.body!;
  }

  static Future<Map<String, dynamic>> getTriggerConfig(String name) async {
    final result = await service.getTriggerConfig(name);
    return result.body!;
  }
}
