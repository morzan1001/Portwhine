part of 'pipeline_errors_bloc.dart';

abstract class PipelineErrorsEvent extends Equatable {
  const PipelineErrorsEvent();

  @override
  List<Object> get props => [];
}

class GetPipelineErrors extends PipelineErrorsEvent {}
