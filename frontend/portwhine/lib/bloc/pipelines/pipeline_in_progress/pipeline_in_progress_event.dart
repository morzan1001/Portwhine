part of 'pipeline_in_progress_bloc.dart';

abstract class PipelineInProgressEvent extends Equatable {
  const PipelineInProgressEvent();

  @override
  List<Object> get props => [];
}

class GetPipelineInProgress extends PipelineInProgressEvent {}
