import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_size_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_status/pipelines_status_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/show_nodes/show_nodes_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/single_pipeline/single_pipeline_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/triggers_list/triggers_list_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/workers_list/workers_list_bloc.dart';

import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';

class BlocProviders {
  static final List<BlocProvider> providers = [
    // pipelines
    BlocProvider<PipelinesListBloc>(
      create: (context) => PipelinesListBloc()..add(const GetAllPipelines()),
    ),
    BlocProvider<PipelinesStatusBloc>(
      create: (context) => PipelinesStatusBloc(),
    ),
    BlocProvider<CreatePipelineBloc>(
      create: (context) => CreatePipelineBloc(),
    ),
    BlocProvider<DeletePipelineBloc>(
      create: (context) => DeletePipelineBloc(),
    ),
    BlocProvider<StartStopPipelineBloc>(
      create: (context) => StartStopPipelineBloc(),
    ),
    BlocProvider<PipelinePageCubit>(
      create: (context) => PipelinePageCubit(),
    ),
    BlocProvider<PipelineSizeCubit>(
      create: (context) => PipelineSizeCubit(),
    ),

    // single pipeline
    BlocProvider<SinglePipelineBloc>(
      create: (context) => SinglePipelineBloc(),
    ),
    BlocProvider<WorkersListBloc>(
      create: (context) => WorkersListBloc(),
    ),
    BlocProvider<TriggersListBloc>(
      create: (context) => TriggersListBloc(),
    ),
    BlocProvider<PipelineCubit>(
      create: (context) => PipelineCubit(),
    ),
    BlocProvider<ShowNodesCubit>(
      create: (context) => ShowNodesCubit(),
    ),

    //
    BlocProvider<NodesCubit>(
      create: (context) => NodesCubit(),
    ),
    BlocProvider<LinesCubit>(
      create: (context) => LinesCubit(),
    ),
    BlocProvider<SelectedNodeCubit>(
      create: (context) => SelectedNodeCubit(),
    ),
    BlocProvider<ConnectingLineCubit>(
      create: (context) => ConnectingLineCubit(),
    ),
    BlocProvider<CanvasCubit>(
      create: (context) => CanvasCubit(),
    ),
  ];
}
