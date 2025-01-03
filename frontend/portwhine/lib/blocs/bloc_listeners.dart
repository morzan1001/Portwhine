import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/get_all_pipelines/get_all_pipelines_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/widgets/toast.dart';

class BlocListeners {
  static final List<BlocListener> pipelinesListener = [
    BlocListener<PipelinePageCubit, int>(
      listener: (context, state) {
        BlocProvider.of<GetAllPipelinesBloc>(context).add(
          GetAllPipelines(page: state),
        );
      },
    ),
    BlocListener<CreatePipelineBloc, CreatePipelineState>(
      listener: (context, state) async {
        if (state is CreatePipelineFailed) {
          showToast(context, state.error);
        }

        if (state is CreatePipelineCompleted) {
          pop(context);
          // todo: added delay so backend gets updated with new data
          await Future.delayed(const Duration(milliseconds: 500));
          BlocProvider.of<GetAllPipelinesBloc>(context).add(
            const GetAllPipelines(),
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
