part of 'pipeline_errors_bloc.dart';

abstract class PipelineErrorsState extends Equatable {
  const PipelineErrorsState();

  @override
  List<Object> get props => [];
}

class PipelineErrorsInitial extends PipelineErrorsState {}

class PipelineErrorsLoading extends PipelineErrorsState {}

class PipelineErrorsLoaded extends PipelineErrorsState {
  final List<String> errors;

  const PipelineErrorsLoaded(this.errors);
}

class PipelineErrorsFailed extends PipelineErrorsState {
  final String error;

  const PipelineErrorsFailed(this.error);
}
