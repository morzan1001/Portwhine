part of 'delete_pipeline_bloc.dart';

abstract class DeletePipelineEvent extends Equatable {
  const DeletePipelineEvent();

  @override
  List<Object> get props => [];
}

class DeletePipeline extends DeletePipelineEvent {
  final String id;

  const DeletePipeline(this.id);
}
