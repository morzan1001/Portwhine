// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// AutoRouterGenerator
// **************************************************************************

// ignore_for_file: type=lint
// coverage:ignore-file

part of 'router.dart';

/// generated route for
/// [PipelineDetailsPage]
class PipelineDetailsRoute extends PageRouteInfo<PipelineDetailsRouteArgs> {
  PipelineDetailsRoute({
    required String id,
    PipelineModel? model,
    Key? key,
    List<PageRouteInfo>? children,
  }) : super(
          PipelineDetailsRoute.name,
          args: PipelineDetailsRouteArgs(
            id: id,
            model: model,
            key: key,
          ),
          rawPathParams: {'id': id},
          initialChildren: children,
        );

  static const String name = 'PipelineDetailsRoute';

  static PageInfo page = PageInfo(
    name,
    builder: (data) {
      final pathParams = data.inheritedPathParams;
      final args = data.argsAs<PipelineDetailsRouteArgs>(
          orElse: () =>
              PipelineDetailsRouteArgs(id: pathParams.getString('id')));
      return PipelineDetailsPage(
        id: args.id,
        model: args.model,
        key: args.key,
      );
    },
  );
}

class PipelineDetailsRouteArgs {
  const PipelineDetailsRouteArgs({
    required this.id,
    this.model,
    this.key,
  });

  final String id;

  final PipelineModel? model;

  final Key? key;

  @override
  String toString() {
    return 'PipelineDetailsRouteArgs{id: $id, model: $model, key: $key}';
  }
}

/// generated route for
/// [PipelinesPage]
class PipelinesRoute extends PageRouteInfo<void> {
  const PipelinesRoute({List<PageRouteInfo>? children})
      : super(
          PipelinesRoute.name,
          initialChildren: children,
        );

  static const String name = 'PipelinesRoute';

  static PageInfo page = PageInfo(
    name,
    builder: (data) {
      return const PipelinesPage();
    },
  );
}
