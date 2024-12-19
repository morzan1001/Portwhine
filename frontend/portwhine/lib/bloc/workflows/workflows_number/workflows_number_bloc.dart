import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:frontend/repos/workflows/workflows_repo.dart';

part 'workflows_number_event.dart';
part 'workflows_number_state.dart';

class PipelinesNumberBloc
    extends Bloc<WorkflowsNumberEvent, WorkflowsNumberState> {
  PipelinesNumberBloc() : super(WorkflowsNumberInitial()) {
    on<GetWorkflowsNumber>(
      (event, emit) async {
        try {
          emit(WorkflowsNumberLoading());
          final result = await WorkflowsRepo().getWorkflowsNumber();
          emit(WorkflowsNumberLoaded(result));
        } catch (e) {
          emit(WorkflowsNumberFailed(e.toString()));
        }
      },
    );
  }
}
