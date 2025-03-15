part of 'workers_list_bloc.dart';

abstract class WorkersListState extends Equatable {
  const WorkersListState();

  @override
  List<Object> get props => [];
}

class WorkersListInitial extends WorkersListState {}

class WorkersListLoading extends WorkersListState {}

class WorkersListLoaded extends WorkersListState {
  final List<String> workers;

  const WorkersListLoaded(this.workers);
}

class WorkersListFailed extends WorkersListState {
  final String error;

  const WorkersListFailed(this.error);
}
