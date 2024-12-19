part of 'workflows_list_bloc.dart';

abstract class WorkflowsListEvent extends Equatable {
  const WorkflowsListEvent();

  @override
  List<Object> get props => [];
}

class GetPipelinesList extends WorkflowsListEvent {}
