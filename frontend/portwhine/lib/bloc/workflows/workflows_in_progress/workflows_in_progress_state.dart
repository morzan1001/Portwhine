part of 'workflows_in_progress_bloc.dart';

abstract class PipelinesInProgressState extends Equatable {
  const PipelinesInProgressState();

  @override
  List<Object> get props => [];
}

class PipelinesInProgressInitial extends PipelinesInProgressState {}

class PipelinesInProgressLoading extends PipelinesInProgressState {}

class PipelinesInProgressLoaded extends PipelinesInProgressState {
  final int number;

  const PipelinesInProgressLoaded(this.number);
}

class PipelinesInProgressFailed extends PipelinesInProgressState {
  final String error;

  const PipelinesInProgressFailed(this.error);
}
