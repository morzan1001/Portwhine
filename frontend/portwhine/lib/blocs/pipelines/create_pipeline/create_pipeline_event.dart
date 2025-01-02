part of 'create_pipeline_bloc.dart';

abstract class CreatePipelineEvent extends Equatable {
  const CreatePipelineEvent();

  @override
  List<Object> get props => [];
}

class CreatePipeline extends CreatePipelineEvent {
  final String name;

  const CreatePipeline(this.name);
}
