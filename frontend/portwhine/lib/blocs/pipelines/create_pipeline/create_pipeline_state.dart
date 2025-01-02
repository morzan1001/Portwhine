part of 'create_pipeline_bloc.dart';

abstract class CreatePipelineState extends Equatable {
  const CreatePipelineState();

  @override
  List<Object> get props => [];
}

class CreatePipelineInitial extends CreatePipelineState {}

class CreatePipelineStarted extends CreatePipelineState {}

class CreatePipelineCompleted extends CreatePipelineState {
  final PipelineModel pipeline;

  const CreatePipelineCompleted(this.pipeline);
}

class CreatePipelineFailed extends CreatePipelineState {
  final String error;

  const CreatePipelineFailed(this.error);
}
