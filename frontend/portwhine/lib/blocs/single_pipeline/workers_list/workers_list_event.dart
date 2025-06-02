part of 'workers_list_bloc.dart';

abstract class WorkersListEvent extends Equatable {
  const WorkersListEvent();

  @override
  List<Object> get props => [];
}

class GetWorkersList extends WorkersListEvent {}
