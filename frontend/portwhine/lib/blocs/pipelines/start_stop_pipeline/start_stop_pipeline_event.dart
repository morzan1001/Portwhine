part of 'start_stop_pipeline_bloc.dart';

abstract class StartStopPipelineEvent extends Equatable {
  const StartStopPipelineEvent();

  @override
  List<Object> get props => [];
}

class StartPipeline extends StartStopPipelineEvent {
  final String id;

  const StartPipeline(this.id);
}

class StopPipeline extends StartStopPipelineEvent {
  final String id;

  const StopPipeline(this.id);
}
