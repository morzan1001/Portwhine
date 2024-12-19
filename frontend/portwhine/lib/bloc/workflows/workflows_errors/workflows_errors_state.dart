part of 'workflows_errors_bloc.dart';

abstract class PipelinesErrorsState extends Equatable {
  const PipelinesErrorsState();

  @override
  List<Object> get props => [];
}

class PipelinesErrorsInitial extends PipelinesErrorsState {}

class PipelinesErrorsLoading extends PipelinesErrorsState {}

class PipelinesErrorsLoaded extends PipelinesErrorsState {
  final List<String> errors;

  const PipelinesErrorsLoaded(this.errors);
}

class PipelinesErrorsFailed extends PipelinesErrorsState {
  final String error;

  const PipelinesErrorsFailed(this.error);
}
