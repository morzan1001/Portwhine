part of 'triggers_list_bloc.dart';

abstract class TriggersListState extends Equatable {
  const TriggersListState();

  @override
  List<Object> get props => [];
}

class TriggersListInitial extends TriggersListState {}

class TriggersListLoading extends TriggersListState {}

class TriggersListLoaded extends TriggersListState {
  final List<String> triggers;

  const TriggersListLoaded(this.triggers);
}

class TriggersListFailed extends TriggersListState {
  final String error;

  const TriggersListFailed(this.error);
}
