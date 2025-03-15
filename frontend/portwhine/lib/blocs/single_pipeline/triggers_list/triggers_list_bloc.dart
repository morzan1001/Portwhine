import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repos/single_pipeline/single_pipeline.dart';

part 'triggers_list_event.dart';
part 'triggers_list_state.dart';

class TriggersListBloc extends Bloc<TriggersListEvent, TriggersListState> {
  TriggersListBloc() : super(TriggersListInitial()) {
    on<GetTriggersList>(
      (event, emit) async {
        try {
          emit(TriggersListLoading());
          final triggers = await SinglePipelineRepo.getAllTriggers();
          emit(TriggersListLoaded(triggers));
        } catch (e) {
          emit(TriggersListFailed(e.toString()));
        }
      },
    );
  }
}
