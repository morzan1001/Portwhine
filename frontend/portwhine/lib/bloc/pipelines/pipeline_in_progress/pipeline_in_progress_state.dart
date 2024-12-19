part of 'pipeline_in_progress_bloc.dart';

abstract class PipelineInProgressState extends Equatable {
  const PipelineInProgressState();

  @override
  List<Object> get props => [];
}

class PipelineInProgressInitial extends PipelineInProgressState {}

class PipelineInProgressLoading extends PipelineInProgressState {}

class PipelineInProgressLoaded extends PipelineInProgressState {
  final int number;

  const PipelineInProgressLoaded(this.number);
}

class PipelineInProgressFailed extends PipelineInProgressState {
  final String error;

  const PipelineInProgressFailed(this.error);
}
