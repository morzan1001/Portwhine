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
    return result.body?['message'] == kPipelineDeleteSuccessMessage;
  }
}
