part of 'workflows_list_bloc.dart';

abstract class PipelinesListState extends Equatable {
  const PipelinesListState();

  @override
  List<Object> get props => [];
}

class PipelinesListInitial extends PipelinesListState {}

class PipelinesListLoading extends PipelinesListState {}

class PipelinesListLoaded extends PipelinesListState {
  final List<WorkflowModel> workflows;

  const PipelinesListLoaded(this.workflows);
}

class PipelinesListFailed extends PipelinesListState {
  final String error;

  const PipelinesListFailed(this.error);
}
