part of 'pipelines_list_bloc.dart';

abstract class PipelinesListEvent extends Equatable {
  const PipelinesListEvent();

  @override
  List<Object> get props => [];
}

class GetAllPipelines extends PipelinesListEvent {
  final int page;
  final int size;

  const GetAllPipelines({this.page = 1, this.size = 10});
}

class DeletePipelineFromList extends PipelinesListEvent {
  final String id;

  const DeletePipelineFromList(this.id);
}
