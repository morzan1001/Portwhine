import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/repos/pipelines/pipelines_repo.dart';

part 'get_all_pipelines_event.dart';
part 'get_all_pipelines_state.dart';

class GetAllPipelinesBloc
    extends Bloc<GetAllPipelinesEvent, GetAllPipelinesState> {
  final repo = PipelinesRepo();

  List<PipelineModel> pipelines = [];

  GetAllPipelinesBloc() : super(GetAllPipelinesInitial()) {
    on<GetAllPipelines>(
      (event, emit) async {
        try {
          emit(GetAllPipelinesLoading());
          final newPipelines = await repo.getAllPipelines(
            page: event.page,
            size: event.size,
          );
          pipelines.clear();
          pipelines.addAll(newPipelines);
          emit(GetAllPipelinesLoaded(pipelines));
        } catch (e) {
          emit(GetAllPipelinesFailed(e.toString()));
        }
      },
    );

    on<DeletePipelineFromList>(
      (event, emit) {
        emit(GetAllPipelinesLoading());
        pipelines.removeWhere((e) => e.id == event.id);
        emit(GetAllPipelinesLoaded(List.from(pipelines)));
      },
    );
  }
}
