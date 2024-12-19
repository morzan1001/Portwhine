import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:frontend/repos/workflows/workflows_repo.dart';

part 'workflows_errors_event.dart';
part 'workflows_errors_state.dart';

class PipelinesErrorsBloc
    extends Bloc<WorkflowsErrorsEvent, WorkflowsErrorsState> {
  PipelinesErrorsBloc() : super(WorkflowsErrorsInitial()) {
    on<GetWorkflowsErrors>(
      (event, emit) async {
        try {
          emit(WorkflowsErrorsLoading());
          final result = await WorkflowsRepo().getWorkflowsErrors();
          emit(WorkflowsErrorsLoaded(result));
        } catch (e) {
          emit(WorkflowsErrorsFailed(e.toString()));
        }
      },
    );
  }
}
