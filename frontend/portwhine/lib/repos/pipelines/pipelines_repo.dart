import 'package:portwhine/api/api.dart';
import 'package:portwhine/models/pipeline_model.dart';

class PipelinesRepo {
  static Future<List<PipelineModel>> getAllPipelines({
    int size = 10,
    int page = 1,
  }) async {
    final pipelines = await Api.getAllPipelines(size: size, page: page);
    return pipelines;
  }

  static Future<PipelineModel> createPipeline(String name) async {
    final result = await Api.createPipeline(name);
    return result;
  }

  static Future<bool> deletePipeline(String id) async {
    final result = await Api.deletePipeline(id);
    return result;
  }

  static Future<Map<String, dynamic>> startPipeline(String id) async {
    final result = await Api.startPipeline(id);
    return result;
  }

  static Future<Map<String, dynamic>> stopPipeline(String id) async {
    final result = await Api.stopPipeline(id);
    return result;
  }
}
