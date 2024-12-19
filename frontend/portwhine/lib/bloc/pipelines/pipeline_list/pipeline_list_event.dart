part of 'pipeline_list_bloc.dart';

abstract class PipelineListEvent extends Equatable {
  const PipelineListEvent();

  @override
  List<Object> get props => [];
}

class GetPipelineList extends PipelineListEvent {}
