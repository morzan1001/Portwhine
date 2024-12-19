part of 'workflows_in_progress_bloc.dart';

abstract class PipelinesInProgressEvent extends Equatable {
  const PipelinesInProgressEvent();

  @override
  List<Object> get props => [];
}

class GetPipelinesInProgress extends PipelinesInProgressEvent {}
