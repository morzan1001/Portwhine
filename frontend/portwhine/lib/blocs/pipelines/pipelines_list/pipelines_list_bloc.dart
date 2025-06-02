import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'pipelines_list_event.dart';
part 'pipelines_list_state.dart';

class PipelinesListBloc extends Bloc<PipelinesListEvent, PipelinesListState> {
  List<PipelineModel> pipelines = [];

  PipelinesListBloc() : super(PipelinesListInitial()) {
    on<GetAllPipelines>(
      (event, emit) async {
        try {
          emit(PipelinesListLoading());
          final newPipelines = await PipelinesRepo.getAllPipelines(
            page: event.page,
            size: event.size,
          );
          pipelines.clear();
          pipelines.addAll(newPipelines);
          emit(PipelinesListLoaded(pipelines));
        } catch (e) {
          emit(PipelinesListFailed(e.toString()));
        }
      },
    );

    on<DeletePipelineFromList>(
      (event, emit) {
        emit(PipelinesListLoading());
        pipelines.removeWhere((e) => e.id == event.id);
        emit(PipelinesListLoaded(List.from(pipelines)));
      },
    );
  }
}
