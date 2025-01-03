part of 'get_all_pipelines_bloc.dart';

abstract class GetAllPipelinesEvent extends Equatable {
  const GetAllPipelinesEvent();

  @override
  List<Object> get props => [];
}

class GetAllPipelines extends GetAllPipelinesEvent {
  final int page;
  final int size;

  const GetAllPipelines({this.page = 1, this.size = 10});
}

class DeletePipelineFromList extends GetAllPipelinesEvent {
  final String id;

  const DeletePipelineFromList(this.id);
}
