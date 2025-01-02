part of 'delete_pipeline_bloc.dart';

abstract class DeletePipelineState extends Equatable {
  const DeletePipelineState();

  @override
  List<Object> get props => [];
}

class DeletePipelineInitial extends DeletePipelineState {}

class DeletePipelineStarted extends DeletePipelineState {
  final String id;

  const DeletePipelineStarted(this.id);
}

class DeletePipelineCompleted extends DeletePipelineState {
  final String id;

  const DeletePipelineCompleted(this.id);
}

class DeletePipelineFailed extends DeletePipelineState {
  final String error;

  const DeletePipelineFailed(this.error);
}
