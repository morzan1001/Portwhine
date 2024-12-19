part of 'workflows_number_bloc.dart';

abstract class PipelinesNumberEvent extends Equatable {
  const PipelinesNumberEvent();

  @override
  List<Object> get props => [];
}

class GetPipelinesNumber extends PipelinesNumberEvent {}
