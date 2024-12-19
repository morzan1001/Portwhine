import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repositories/pipeline/pipeline_repo.dart';

part 'pipeline_errors_event.dart';
part 'pipeline_errors_state.dart';

class PipelineErrorsBloc
    extends Bloc<PipelineErrorsEvent, PipelineErrorsState> {
  PipelineErrorsBloc() : super(PipelineErrorsInitial()) {
    on<GetPipelineErrors>(
      (event, emit) async {
        try {
          emit(PipelineErrorsLoading());
          final result = await PipelineRepo().getPipelineErrors();
          emit(PipelineErrorsLoaded(result));
        } catch (e) {
          emit(PipelineErrorsFailed(e.toString()));
        }
      },
    );
  }
}
