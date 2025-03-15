part of 'single_pipeline_bloc.dart';

abstract class SinglePipelineEvent extends Equatable {
  const SinglePipelineEvent();

  @override
  List<Object> get props => [];
}

class GetSinglePipeline extends SinglePipelineEvent {
  final String id;

  const GetSinglePipeline(this.id);
}
