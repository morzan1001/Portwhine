part of 'workflows_errors_bloc.dart';

abstract class WorkflowsErrorsEvent extends Equatable {
  const WorkflowsErrorsEvent();

  @override
  List<Object> get props => [];
}

class GetWorkflowsErrors extends WorkflowsErrorsEvent {}
