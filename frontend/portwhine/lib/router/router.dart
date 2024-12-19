import 'package:auto_route/auto_route.dart';
import 'package:flutter/cupertino.dart';
import 'package:frontend/pages/pipelines/pipelines.dart';
import 'package:frontend/pages/pipelines/sections/list/pipeline_item.dart';

part 'router.gr.dart';

@AutoRouterConfig()
class AppRouter extends RootStackRouter {
  @override
  List<AutoRoute> get routes {
    return [
      CustomRoute(
        path: '/',
        page: PipelinesPage.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
      CustomRoute(
        path: '/pipeline/:id',
        page: PipelineItem.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
    ];
  }
}
