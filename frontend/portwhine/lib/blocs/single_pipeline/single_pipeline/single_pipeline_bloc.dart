import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/repos/single_pipeline/single_pipeline.dart';

part 'single_pipeline_event.dart';
part 'single_pipeline_state.dart';

class SinglePipelineBloc
    extends Bloc<SinglePipelineEvent, SinglePipelineState> {
  SinglePipelineBloc() : super(SinglePipelineInitial()) {
    on<GetSinglePipeline>(
      (event, emit) async {
        try {
          emit(SinglePipelineLoading());
          final pipeline = await SinglePipelineRepo.getPipeline(event.id);
          emit(SinglePipelineLoaded(pipeline));
        } catch (e) {
          emit(SinglePipelineFailed(e.toString()));
        }
      },
    );
  }
}
