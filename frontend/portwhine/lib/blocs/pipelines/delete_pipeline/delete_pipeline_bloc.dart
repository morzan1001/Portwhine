import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'delete_pipeline_event.dart';
part 'delete_pipeline_state.dart';

class DeletePipelineBloc
    extends Bloc<DeletePipelineEvent, DeletePipelineState> {
  DeletePipelineBloc() : super(DeletePipelineInitial()) {
    on<DeletePipeline>(
      (event, emit) async {
        final repo = PipelinesRepo();

        try {
          emit(DeletePipelineStarted(event.id));
          final deleted = await repo.deletePipeline(event.id);
          if (deleted) {
            emit(DeletePipelineCompleted(event.id));
          } else {
            emit(const DeletePipelineFailed('Error occurred'));
          }
        } catch (e) {
          emit(DeletePipelineFailed(e.toString()));
        }
      },
    );
  }
}
