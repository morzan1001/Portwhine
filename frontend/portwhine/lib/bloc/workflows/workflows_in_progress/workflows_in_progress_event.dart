part of 'workflows_in_progress_bloc.dart';

abstract class WorkflowsInProgressEvent extends Equatable {
  const WorkflowsInProgressEvent();

  @override
  List<Object> get props => [];
}

class GetWorkflowsInProgress extends WorkflowsInProgressEvent {}
