part of 'single_pipeline_bloc.dart';

abstract class SinglePipelineState extends Equatable {
  const SinglePipelineState();

  @override
  List<Object> get props => [];
}

class SinglePipelineInitial extends SinglePipelineState {}

class SinglePipelineLoading extends SinglePipelineState {}

class SinglePipelineLoaded extends SinglePipelineState {
  final PipelineModel pipeline;

  const SinglePipelineLoaded(this.pipeline);
}

class SinglePipelineFailed extends SinglePipelineState {
  final String error;

  const SinglePipelineFailed(this.error);
}
