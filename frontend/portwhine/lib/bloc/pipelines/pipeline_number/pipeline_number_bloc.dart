import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repositories/pipeline/pipeline_repo.dart';

part 'pipeline_number_event.dart';
part 'pipeline_number_state.dart';

class PipelineNumberBloc
    extends Bloc<PipelineNumberEvent, PipelineNumberState> {
  PipelineNumberBloc() : super(PipelineNumberInitial()) {
    on<GetPipelineNumber>(
      (event, emit) async {
        try {
          emit(PipelineNumberLoading());
          final result = await PipelineRepo().getPipelineNumber();
          emit(PipelineNumberLoaded(result));
        } catch (e) {
          emit(PipelineNumberFailed(e.toString()));
        }
      },
    );
  }
}
