import 'package:auto_route/auto_route.dart';
import 'package:flutter/cupertino.dart';

part 'router.gr.dart';

@AutoRouterConfig()
class AppRouter extends RootStackRouter {
  @override
  List<AutoRoute> get routes {
    return [
      CustomRoute(
        path: '/',
        page: PipelineList.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
      CustomRoute(
        path: '/pipeline/:id',
        page: PipelineDetail.page,
        transitionsBuilder: TransitionsBuilders.noTransition,
      ),
    ];
  }
}
