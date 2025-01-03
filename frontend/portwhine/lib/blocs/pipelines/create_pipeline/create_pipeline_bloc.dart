import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'create_pipeline_event.dart';
part 'create_pipeline_state.dart';

class CreatePipelineBloc
    extends Bloc<CreatePipelineEvent, CreatePipelineState> {
  CreatePipelineBloc() : super(CreatePipelineInitial()) {
    on<CreatePipeline>(
      (event, emit) async {
        final repo = PipelinesRepo();

        try {
          emit(CreatePipelineStarted());
          final pipeline = await repo.createPipeline(event.name);
          emit(CreatePipelineCompleted(pipeline));
        } catch (e) {
          emit(CreatePipelineFailed(e.toString()));
        }
      },
    );
  }
}
