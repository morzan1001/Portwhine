part of 'pipelines_status_bloc.dart';

abstract class PipelinesStatusEvent extends Equatable {
  const PipelinesStatusEvent();

  @override
  List<Object> get props => [];
}

class UpdatePipelinesList extends PipelinesStatusEvent {
  final List<PipelineModel> pipelines;

  const UpdatePipelinesList(this.pipelines);
}

class UpdatePipelineStatus extends PipelinesStatusEvent {
  final String id;
  final String newStatus;

  const UpdatePipelineStatus(this.id, this.newStatus);
}
