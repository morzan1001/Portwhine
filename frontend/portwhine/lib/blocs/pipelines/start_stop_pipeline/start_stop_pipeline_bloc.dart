import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'start_stop_pipeline_event.dart';
part 'start_stop_pipeline_state.dart';

class StartStopPipelineBloc
    extends Bloc<StartStopPipelineEvent, StartStopPipelineState> {
  StartStopPipelineBloc() : super(StartStopPipelineInitial()) {
    on<StartPipeline>(
      (event, emit) async {
        try {
          emit(StartStopPipelineStarted(event.id));
          final result = await PipelinesRepo.startPipeline(event.id);
          final String message = result['detail'];

          if (message.contains(kPipelineStarted)) {
            emit(StartStopPipelineCompleted(event.id, kStatusRunning, message));
          } else {
            emit(StartStopPipelineFailed(message));
          }
        } catch (e) {
          emit(StartStopPipelineFailed(e.toString()));
        }
      },
    );

    on<StopPipeline>(
      (event, emit) async {
        try {
          emit(StartStopPipelineStarted(event.id));
          final result = await PipelinesRepo.stopPipeline(event.id);
          final String message = result['detail'];

          if (message.contains(kPipelineStopped)) {
            emit(StartStopPipelineCompleted(event.id, kStatusStopped, message));
          } else {
            emit(StartStopPipelineFailed(message));
          }
        } catch (e) {
          emit(StartStopPipelineFailed(e.toString()));
        }
      },
    );
  }
}
