import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/get_all_pipelines/get_all_pipelines_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';

class BlocProviders {
  static final List<BlocProvider> providers = [
    // pipelines
    BlocProvider<GetAllPipelinesBloc>(
      create: (context) => GetAllPipelinesBloc()..add(GetAllPipelines()),
    ),
    BlocProvider<CreatePipelineBloc>(
      create: (context) => CreatePipelineBloc(),
    ),
    BlocProvider<DeletePipelineBloc>(
      create: (context) => DeletePipelineBloc(),
    ),

    // single pipeline
    BlocProvider<PipelineCubit>(
      create: (context) => PipelineCubit(),
    ),
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
