import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repositories/pipeline/pipeline_repo.dart';

part 'pipeline_in_progress_event.dart';
part 'pipeline_in_progress_state.dart';

class PipelinesInProgressBloc
    extends Bloc<PipelineInProgressEvent, PipelineInProgressState> {
  PipelinesInProgressBloc() : super(PipelineInProgressInitial()) {
    on<GetPipelineInProgress>(
      (event, emit) async {
        try {
          emit(PipelineInProgressLoading());
          final result = await PipelineRepo().getPipelineInProgress();
          emit(PipelineInProgressLoaded(result));
        } catch (e) {
          emit(PipelineInProgressFailed(e.toString()));
        }
      },
    );
  }
}
