import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_size_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_status/pipelines_status_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/widgets/toast.dart';

class BlocListeners {
  static final List<BlocListener> pipelinesListener = [
    BlocListener<PipelinesListBloc, PipelinesListState>(
      listener: (context, state) {
        if (state is PipelinesListLoaded) {
          BlocProvider.of<PipelinesStatusBloc>(context).add(
            UpdatePipelinesList(state.pipelines),
          );
        }
      },
    ),
    BlocListener<PipelinePageCubit, int>(
      listener: (context, state) {
        BlocProvider.of<PipelinesListBloc>(context).add(
          GetAllPipelines(page: state),
        );
      },
    ),
    BlocListener<PipelineSizeCubit, int>(
      listener: (context, state) {
        BlocProvider.of<PipelinePageCubit>(context).setFirstPage();
        BlocProvider.of<PipelinesListBloc>(context).add(
          GetAllPipelines(size: state),
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
          BlocProvider.of<PipelinesListBloc>(context).add(
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
          BlocProvider.of<PipelinesListBloc>(context).add(
            DeletePipelineFromList(state.id),
          );
        }
      },
    ),
    BlocListener<StartStopPipelineBloc, StartStopPipelineState>(
      listener: (context, state) {
        if (state is StartStopPipelineFailed) {
          showToast(context, state.error);
        }

        if (state is StartStopPipelineCompleted) {
          showToast(
            context,
            state.message,
            color: switch (state.status) {
              kStatusRunning => MyColors.green,
              kStatusStopped => MyColors.red,
              _ => MyColors.prime,
            },
          );
          BlocProvider.of<PipelinesStatusBloc>(context).add(
            UpdatePipelineStatus(state.id, state.status),
          );
        }
      },
    ),
  ];
}
