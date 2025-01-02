import 'package:portwhine/api/api.dart';
import 'package:portwhine/models/pipeline_model.dart';

class PipelinesRepo {
  Future<List<PipelineModel>> getGetAllPipelines() async {
    final pipelines = await Api.getAllPipelines();
    return pipelines;
  }

  Future<PipelineModel> createPipeline(String name) async {
    final result = await Api.createPipeline(name);
    return result;
  }

  Future<bool> deletePipeline(String id) async {
    final result = await Api.deletePipeline(id);
    return result;
  }
}
