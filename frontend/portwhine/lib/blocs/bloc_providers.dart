import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';

class BlocProviders {
  static final List<BlocProvider> providers = [
    // pipelines list
    BlocProvider<PipelinesListBloc>(
      create: (context) => PipelinesListBloc()
        ..add(
          GetPipelinesList(),
        ),
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
