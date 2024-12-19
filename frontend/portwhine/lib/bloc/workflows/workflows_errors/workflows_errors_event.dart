part of 'workflows_errors_bloc.dart';

abstract class PipelinesErrorsEvent extends Equatable {
  const PipelinesErrorsEvent();

  @override
  List<Object> get props => [];
}

class GetPipelinesErrors extends PipelinesErrorsEvent {}
