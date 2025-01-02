part of 'get_all_pipelines_bloc.dart';

abstract class GetAllPipelinesEvent extends Equatable {
  const GetAllPipelinesEvent();

  @override
  List<Object> get props => [];
}

class GetAllPipelines extends GetAllPipelinesEvent {}

class DeletePipelineFromList extends GetAllPipelinesEvent {
  final String id;

  const DeletePipelineFromList(this.id);
}
