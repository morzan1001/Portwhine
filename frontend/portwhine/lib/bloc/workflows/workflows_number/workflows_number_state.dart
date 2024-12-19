part of 'workflows_number_bloc.dart';

abstract class WorkflowsNumberState extends Equatable {
  const WorkflowsNumberState();

  @override
  List<Object> get props => [];
}

class WorkflowsNumberInitial extends WorkflowsNumberState {}

class WorkflowsNumberLoading extends WorkflowsNumberState {}

class WorkflowsNumberLoaded extends WorkflowsNumberState {
  final Map<DateTime, int> numbers;

  const WorkflowsNumberLoaded(this.numbers);
}

class WorkflowsNumberFailed extends WorkflowsNumberState {
  final String error;

  const WorkflowsNumberFailed(this.error);
}
