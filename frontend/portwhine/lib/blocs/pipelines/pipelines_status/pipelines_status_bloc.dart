import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';

part 'pipelines_status_event.dart';
part 'pipelines_status_state.dart';

class PipelinesStatusBloc
    extends Bloc<PipelinesStatusEvent, PipelinesStatusState> {
  List<PipelineModel> pipelines = [];

  PipelinesStatusBloc() : super(PipelinesStatusInitial()) {
    on<UpdatePipelinesList>(
      (event, emit) {
        emit(PipelinesStatusLoading());
        pipelines = List.from(event.pipelines);
        emit(PipelinesStatusUpdated(pipelines));
      },
    );

    on<UpdatePipelineStatus>(
      (event, emit) {
        final index = pipelines.indexWhere(
          (pipeline) => pipeline.id == event.id,
        );
        if (index != -1) {
          emit(PipelinesStatusLoading());
          pipelines[index] = pipelines[index].copyWith(status: event.newStatus);
          emit(PipelinesStatusUpdated(List.from(pipelines)));
        }
      },
    );
  }
}
