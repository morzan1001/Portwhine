import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/get_all_pipelines/get_all_pipelines_bloc.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/widgets/toast.dart';

class BlocListeners {
  static final List<BlocListener> pipelinesListener = [
    BlocListener<CreatePipelineBloc, CreatePipelineState>(
      listener: (context, state) {
        if (state is CreatePipelineFailed) {
          showToast(context, state.error);
        }

        if (state is CreatePipelineCompleted) {
          pop(context);
          BlocProvider.of<GetAllPipelinesBloc>(context).add(
            GetAllPipelines(),
          );
        }
      },
    ),
    BlocListener<DeletePipelineBloc, DeletePipelineState>(
      listener: (context, state) {
        if (state is DeletePipelineFailed) {
          showToast(context, state.error);
        }

        if (state is DeletePipelineCompleted) {
          BlocProvider.of<GetAllPipelinesBloc>(context).add(
            DeletePipelineFromList(state.id),
          );
        }
      },
    ),
  ];
}
