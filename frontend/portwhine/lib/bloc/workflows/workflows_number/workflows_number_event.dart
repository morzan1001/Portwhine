part of 'workflows_number_bloc.dart';

abstract class WorkflowsNumberEvent extends Equatable {
  const WorkflowsNumberEvent();

  @override
  List<Object> get props => [];
}

class GetWorkflowsNumber extends WorkflowsNumberEvent {}
