part of 'pipeline_number_bloc.dart';

abstract class PipelineNumberEvent extends Equatable {
  const PipelineNumberEvent();

  @override
  List<Object> get props => [];
}

class GetPipelineNumber extends PipelineNumberEvent {}
