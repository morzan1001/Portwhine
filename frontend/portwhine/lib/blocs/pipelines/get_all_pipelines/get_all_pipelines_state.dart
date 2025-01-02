part of 'get_all_pipelines_bloc.dart';

abstract class GetAllPipelinesState extends Equatable {
  const GetAllPipelinesState();

  @override
  List<Object> get props => [];
}

class GetAllPipelinesInitial extends GetAllPipelinesState {}

class GetAllPipelinesLoading extends GetAllPipelinesState {}

class GetAllPipelinesLoaded extends GetAllPipelinesState {
  final List<PipelineModel> pipelines;

  const GetAllPipelinesLoaded(this.pipelines);
}

class GetAllPipelinesFailed extends GetAllPipelinesState {
  final String error;

  const GetAllPipelinesFailed(this.error);
}
