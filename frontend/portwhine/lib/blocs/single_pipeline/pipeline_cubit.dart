import 'package:bloc/bloc.dart';
import 'package:portwhine/models/pipeline_model.dart';

class PipelineCubit extends Cubit<PipelineModel> {
  PipelineCubit() : super(PipelineModel());

  void setPipeline(PipelineModel model) {
    emit(model);
  }
}
