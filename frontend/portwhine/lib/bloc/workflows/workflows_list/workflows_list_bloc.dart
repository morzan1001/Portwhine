import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:frontend/models/workflow_model.dart';
import 'package:frontend/repos/workflows/workflows_repo.dart';

part 'workflows_list_event.dart';
part 'workflows_list_state.dart';

class PipelinesListBloc extends Bloc<WorkflowsListEvent, WorkflowsListState> {
  PipelinesListBloc() : super(WorkflowsListInitial()) {
    on<GetWorkflowsList>(
      (event, emit) async {
        try {
          emit(WorkflowsListLoading());
          final result = await WorkflowsRepo().getWorkflowsList();
          emit(WorkflowsListLoaded(result));
        } catch (e) {
          emit(WorkflowsListFailed(e.toString()));
        }
      },
    );
  }
}
