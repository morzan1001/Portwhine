part of 'workflows_number_bloc.dart';

abstract class PipelinesNumberState extends Equatable {
  const PipelinesNumberState();

  @override
  List<Object> get props => [];
}

class PipelinesNumberInitial extends PipelinesNumberState {}

class PipelinesNumberLoading extends PipelinesNumberState {}

class PipelinesNumberLoaded extends PipelinesNumberState {
  final Map<DateTime, int> numbers;

  const PipelinesNumberLoaded(this.numbers);
}

class PipelinesNumberFailed extends PipelinesNumberState {
  final String error;

  const PipelinesNumberFailed(this.error);
}
