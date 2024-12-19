part of 'pipeline_list_bloc.dart';

abstract class PipelineListState extends Equatable {
  const PipelineListState();

  @override
  List<Object> get props => [];
}

class PipelineListInitial extends PipelineListState {}

class PipelineListLoading extends PipelineListState {}

class PipelineListLoaded extends PipelineListState {
  final List<PipelineModel> pipelines;

  const PipelineListLoaded(this.pipelines);
}

class PipelineListFailed extends PipelineListState {
  final String error;

  const PipelineListFailed(this.error);
}
