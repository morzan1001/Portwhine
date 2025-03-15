import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repos/single_pipeline/single_pipeline.dart';

part 'workers_list_event.dart';
part 'workers_list_state.dart';

class WorkersListBloc extends Bloc<WorkersListEvent, WorkersListState> {
  WorkersListBloc() : super(WorkersListInitial()) {
    on<GetWorkersList>(
      (event, emit) async {
        try {
          emit(WorkersListLoading());
          final workers = await SinglePipelineRepo.getAllWorkers();
          emit(WorkersListLoaded(workers));
        } catch (e) {
          emit(WorkersListFailed(e.toString()));
        }
      },
    );
  }
}
