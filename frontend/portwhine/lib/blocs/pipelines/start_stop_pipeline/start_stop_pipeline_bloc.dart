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
          final started = await PipelinesRepo.startPipeline(event.id);
          if (started) {
            emit(StartStopPipelineCompleted(event.id, kStatusRunning));
          } else {
            emit(const StartStopPipelineFailed('Error occurred'));
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
          final stopped = await PipelinesRepo.stopPipeline(event.id);
          if (stopped) {
            emit(StartStopPipelineCompleted(event.id, kStatusStopped));
          } else {
            emit(const StartStopPipelineFailed('Error occurred'));
          }
        } catch (e) {
          emit(StartStopPipelineFailed(e.toString()));
        }
      },
    );
  }
}
