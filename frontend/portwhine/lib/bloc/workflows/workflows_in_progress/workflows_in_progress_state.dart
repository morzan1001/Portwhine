part of 'workflows_in_progress_bloc.dart';

abstract class WorkflowsInProgressState extends Equatable {
  const WorkflowsInProgressState();

  @override
  List<Object> get props => [];
}

class WorkflowsInProgressInitial extends WorkflowsInProgressState {}

class WorkflowsInProgressLoading extends WorkflowsInProgressState {}

class WorkflowsInProgressLoaded extends WorkflowsInProgressState {
  final int number;

  const WorkflowsInProgressLoaded(this.number);
}

class WorkflowsInProgressFailed extends WorkflowsInProgressState {
  final String error;

  const WorkflowsInProgressFailed(this.error);
}
