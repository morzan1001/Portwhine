part of 'pipeline_number_bloc.dart';

abstract class PipelineNumberState extends Equatable {
  const PipelineNumberState();

  @override
  List<Object> get props => [];
}

class PipelineNumberInitial extends PipelineNumberState {}

class PipelineNumberLoading extends PipelineNumberState {}

class PipelineNumberLoaded extends PipelineNumberState {
  final Map<DateTime, int> numbers;

  const PipelineNumberLoaded(this.numbers);
}

class PipelineNumberFailed extends PipelineNumberState {
  final String error;

  const PipelineNumberFailed(this.error);
}
