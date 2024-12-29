import 'package:auto_route/auto_route.dart';
import 'package:flutter/cupertino.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/pages/pipeline_details/pipeline_details.dart';
import 'package:portwhine/pages/pipelines/pipelines.dart';

part 'router.gr.dart';

@AutoRouterConfig()
class AppRouter extends RootStackRouter {
  @override
  List<AutoRoute> get routes {
    return [
      RedirectRoute(path: '/', redirectTo: '/pipelines'),
      CustomRoute(
        path: '/pipelines',
        page: PipelinesRoute.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
      CustomRoute(
        path: '/pipelines/:id',
        page: PipelineDetailsRoute.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
    ];
  }
}
