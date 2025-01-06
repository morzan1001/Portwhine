part of 'start_stop_pipeline_bloc.dart';

abstract class StartStopPipelineState extends Equatable {
  const StartStopPipelineState();

  @override
  List<Object> get props => [];
}

class StartStopPipelineInitial extends StartStopPipelineState {}

class StartStopPipelineStarted extends StartStopPipelineState {
  final String id;

  const StartStopPipelineStarted(this.id);
}

class StartStopPipelineCompleted extends StartStopPipelineState {
  final String id;
  final String status;

  const StartStopPipelineCompleted(this.id, this.status);
}

class StartStopPipelineFailed extends StartStopPipelineState {
  final String error;

  const StartStopPipelineFailed(this.error);
}
