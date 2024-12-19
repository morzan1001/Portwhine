part of 'workflows_list_bloc.dart';

abstract class WorkflowsListState extends Equatable {
  const WorkflowsListState();

  @override
  List<Object> get props => [];
}

class WorkflowsListInitial extends WorkflowsListState {}

class WorkflowsListLoading extends WorkflowsListState {}

class WorkflowsListLoaded extends WorkflowsListState {
  final List<WorkflowModel> workflows;

  const WorkflowsListLoaded(this.workflows);
}

class WorkflowsListFailed extends WorkflowsListState {
  final String error;

  const WorkflowsListFailed(this.error);
}
