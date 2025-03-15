part of 'triggers_list_bloc.dart';

abstract class TriggersListEvent extends Equatable {
  const TriggersListEvent();

  @override
  List<Object> get props => [];
}

class GetTriggersList extends TriggersListEvent {}
