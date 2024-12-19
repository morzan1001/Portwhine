import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:frontend/repos/workflows/workflows_repo.dart';

part 'workflows_in_progress_event.dart';
part 'workflows_in_progress_state.dart';

class PipelinesInProgressBloc
    extends Bloc<WorkflowsInProgressEvent, WorkflowsInProgressState> {
  PipelinesInProgressBloc() : super(WorkflowsInProgressInitial()) {
    on<GetWorkflowsInProgress>(
      (event, emit) async {
        try {
          emit(WorkflowsInProgressLoading());
          final result = await WorkflowsRepo().getWorkflowsInProgress();
          emit(WorkflowsInProgressLoaded(result));
        } catch (e) {
          emit(WorkflowsInProgressFailed(e.toString()));
        }
      },
    );
  }
}
