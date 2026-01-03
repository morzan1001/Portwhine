import 'package:portwhine/api/api.dart' as gen;
import 'package:portwhine/models/pipeline_model.dart';

class PipelinesRepo {
  static Future<List<PipelineModel>> getAllPipelines({
    int size = 10,
    int page = 1,
  }) async {
    final response = await gen.api.apiV1PipelinesGet(size: size, page: page);
    if (!response.isSuccessful || response.body == null) {
      throw Exception('Failed to load pipelines: ${response.statusCode}');
    }
    // response.body is already List<PipelineListItem> parsed by Chopper
    return response.body!
        .map(
          (item) => PipelineModel(
            id: item.id,
            name: item.name,
            status: item.status ?? 'Unknown',
          ),
        )
        .toList();
  }

  static Future<PipelineModel> createPipeline(String name) async {
    final response = await gen.api.apiV1PipelinePost(
      body: gen.Pipeline(
        name: name,
        trigger: null, // No trigger initially - can be added later
        worker: const [],
      ),
    );
    if (!response.isSuccessful || response.body == null) {
      throw Exception('Failed to create pipeline: ${response.statusCode}');
    }
    // response.body is already PipelineResponse parsed by Chopper
    final body = response.body!;
    return PipelineModel(
      id: body.id,
      name: body.name,
      status: body.status ?? 'Unknown',
    );
  }

  static Future<bool> deletePipeline(String id) async {
    final response =
        await gen.api.apiV1PipelinePipelineIdDelete(pipelineId: id);
    return response.isSuccessful;
  }

  static Future<Map<String, dynamic>> startPipeline(String id) async {
    final response =
        await gen.api.apiV1PipelineStartPipelineIdPost(pipelineId: id);
    if (!response.isSuccessful) {
      throw Exception('Failed to start pipeline: ${response.statusCode}');
    }
    return {'detail': 'Pipeline started successfully'};
  }

  static Future<Map<String, dynamic>> stopPipeline(String id) async {
    final response =
        await gen.api.apiV1PipelineStopPipelineIdPost(pipelineId: id);
    if (!response.isSuccessful) {
      throw Exception('Failed to stop pipeline: ${response.statusCode}');
    }
    return {'detail': 'Pipeline stopped successfully'};
  }
}
