part of 'pipelines_list_bloc.dart';

abstract class PipelinesListEvent extends Equatable {
  const PipelinesListEvent();

  @override
  List<Object> get props => [];
}

class GetPipelinesList extends PipelinesListEvent {}
