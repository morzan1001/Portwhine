part of 'pipelines_status_bloc.dart';

abstract class PipelinesStatusState extends Equatable {
  const PipelinesStatusState();

  @override
  List<Object> get props => [];
}

class PipelinesStatusInitial extends PipelinesStatusState {}

class PipelinesStatusLoading extends PipelinesStatusState {}

class PipelinesStatusUpdated extends PipelinesStatusState {
  final List<PipelineModel> pipelines;

  const PipelinesStatusUpdated(this.pipelines);
}
