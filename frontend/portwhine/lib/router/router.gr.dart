// dart format width=80
// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// AutoRouterGenerator
// **************************************************************************

// ignore_for_file: type=lint
// coverage:ignore-file

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:auto_route/auto_route.dart' as _i4;
import 'package:flutter/material.dart' as _i5;
import 'package:portwhine/pages/pipeline_details/pipeline_details.dart' as _i2;
import 'package:portwhine/pages/pipelines/pipelines.dart' as _i3;
import 'package:portwhine/pages/test/pd_test.dart' as _i1;

/// generated route for
/// [_i1.PDTestPage]
class PDTestRoute extends _i4.PageRouteInfo<void> {
  const PDTestRoute({List<_i4.PageRouteInfo>? children})
    : super(PDTestRoute.name, initialChildren: children);

  static const String name = 'PDTestRoute';

  static _i4.PageInfo page = _i4.PageInfo(
    name,
    builder: (data) {
      return const _i1.PDTestPage();
    },
  );
}

/// generated route for
/// [_i2.PipelineDetailsPage]
class PipelineDetailsRoute extends _i4.PageRouteInfo<PipelineDetailsRouteArgs> {
  PipelineDetailsRoute({
    required String id,
    _i5.Key? key,
    List<_i4.PageRouteInfo>? children,
  }) : super(
         PipelineDetailsRoute.name,
         args: PipelineDetailsRouteArgs(id: id, key: key),
         rawPathParams: {'id': id},
         initialChildren: children,
       );

  static const String name = 'PipelineDetailsRoute';

  static _i4.PageInfo page = _i4.PageInfo(
    name,
    builder: (data) {
      final pathParams = data.inheritedPathParams;
      final args = data.argsAs<PipelineDetailsRouteArgs>(
        orElse: () => PipelineDetailsRouteArgs(id: pathParams.getString('id')),
      );
      return _i2.PipelineDetailsPage(id: args.id, key: args.key);
    },
  );
}

class PipelineDetailsRouteArgs {
  const PipelineDetailsRouteArgs({required this.id, this.key});

  final String id;

  final _i5.Key? key;

  @override
  String toString() {
    return 'PipelineDetailsRouteArgs{id: $id, key: $key}';
  }
}

/// generated route for
/// [_i3.PipelinesPage]
class PipelinesRoute extends _i4.PageRouteInfo<void> {
  const PipelinesRoute({List<_i4.PageRouteInfo>? children})
    : super(PipelinesRoute.name, initialChildren: children);

  static const String name = 'PipelinesRoute';

  static _i4.PageInfo page = _i4.PageInfo(
    name,
    builder: (data) {
      return const _i3.PipelinesPage();
    },
  );
}
