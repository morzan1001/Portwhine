import 'package:portwhine/api/api.dart';
import 'package:portwhine/models/pipeline_model.dart';

class SinglePipelineRepo {
  static Future<PipelineModel> getPipeline(String id) async {
    final result = await Api.getPipeline(id);
    return result;
  }

  static Future<List<String>> getAllWorkers() async {
    final result = await Api.getAllWorkers();
    return result;
  }

  static Future<List<String>> getAllTriggers() async {
    final result = await Api.getAllTriggers();
    return result;
  }
}
