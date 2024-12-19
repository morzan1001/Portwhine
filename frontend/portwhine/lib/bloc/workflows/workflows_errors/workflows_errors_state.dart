part of 'workflows_errors_bloc.dart';

abstract class WorkflowsErrorsState extends Equatable {
  const WorkflowsErrorsState();

  @override
  List<Object> get props => [];
}

class WorkflowsErrorsInitial extends WorkflowsErrorsState {}

class WorkflowsErrorsLoading extends WorkflowsErrorsState {}

class WorkflowsErrorsLoaded extends WorkflowsErrorsState {
  final List<String> errors;

  const WorkflowsErrorsLoaded(this.errors);
}

class WorkflowsErrorsFailed extends WorkflowsErrorsState {
  final String error;

  const WorkflowsErrorsFailed(this.error);
}
