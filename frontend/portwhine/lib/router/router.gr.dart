// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// AutoRouterGenerator
// **************************************************************************

// ignore_for_file: type=lint
// coverage:ignore-file

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:auto_route/auto_route.dart' as _i3;
import 'package:flutter/material.dart' as _i5;
import 'package:portwhine/models/pipeline_model.dart' as _i4;
import 'package:portwhine/pages/pipeline_details/pipeline_details.dart' as _i1;
import 'package:portwhine/pages/pipelines/pipelines.dart' as _i2;

/// generated route for
/// [_i1.PipelineDetailsPage]
class PipelineDetailsRoute extends _i3.PageRouteInfo<PipelineDetailsRouteArgs> {
  PipelineDetailsRoute({
    required String id,
    _i4.PipelineModel? model,
    _i5.Key? key,
    List<_i3.PageRouteInfo>? children,
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

  static _i3.PageInfo page = _i3.PageInfo(
    name,
    builder: (data) {
      final pathParams = data.inheritedPathParams;
      final args = data.argsAs<PipelineDetailsRouteArgs>(
          orElse: () =>
              PipelineDetailsRouteArgs(id: pathParams.getString('id')));
      return _i1.PipelineDetailsPage(
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

  final _i4.PipelineModel? model;

  final _i5.Key? key;

  @override
  String toString() {
    return 'PipelineDetailsRouteArgs{id: $id, model: $model, key: $key}';
  }
}

/// generated route for
/// [_i2.PipelinesPage]
class PipelinesRoute extends _i3.PageRouteInfo<void> {
  const PipelinesRoute({List<_i3.PageRouteInfo>? children})
      : super(
          PipelinesRoute.name,
          initialChildren: children,
        );

  static const String name = 'PipelinesRoute';

  static _i3.PageInfo page = _i3.PageInfo(
    name,
    builder: (data) {
      return const _i2.PipelinesPage();
    },
  );
}
