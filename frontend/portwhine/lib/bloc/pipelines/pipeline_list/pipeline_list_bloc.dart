import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/repositories/pipeline/pipeline_repo.dart';
import 'package:portwhine/models/pipeline_model.dart';

part 'pipeline_list_event.dart';
part 'pipeline_list_state.dart';

class PipelineListBloc extends Bloc<PipelineListEvent, PipelineListState> {
  PipelineListBloc() : super(PipelineListInitial()) {
    on<GetPipelineList>(
      (event, emit) async {
        try {
          emit(PipelineListLoading());
          final result = await PipelineRepo().getPipelineList();
          emit(PipelineListLoaded(result));
        } catch (e) {
          emit(PipelineListFailed(e.toString()));
        }
      },
    );
  }
}
