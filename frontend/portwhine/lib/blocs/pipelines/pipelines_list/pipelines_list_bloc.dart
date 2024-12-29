import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'pipelines_list_event.dart';
part 'pipelines_list_state.dart';

class PipelinesListBloc extends Bloc<PipelinesListEvent, PipelinesListState> {
  PipelinesListBloc() : super(PipelinesListInitial()) {
    on<GetPipelinesList>(
      (event, emit) async {
        try {
          emit(PipelinesListLoading());
          final result = await PipelinesRepo().getPipelinesList();
          emit(PipelinesListLoaded(result));
        } catch (e) {
          emit(PipelinesListFailed(e.toString()));
        }
      },
    );
  }
}
