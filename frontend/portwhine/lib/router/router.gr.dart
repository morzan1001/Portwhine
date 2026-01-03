// dart format width=80
// GENERATED CODE - DO NOT MODIFY BY HAND

// **************************************************************************
// AutoRouterGenerator
// **************************************************************************

// ignore_for_file: type=lint
// coverage:ignore-file

// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:auto_route/auto_route.dart' as _i5;
import 'package:flutter/material.dart' as _i6;
import 'package:portwhine/pages/pipeline_details/pipeline_details.dart' as _i2;
import 'package:portwhine/pages/pipeline_details/results_page.dart' as _i4;
import 'package:portwhine/pages/pipelines/pipelines.dart' as _i3;
import 'package:portwhine/pages/test/pd_test.dart' as _i1;

/// generated route for
/// [_i1.PDTestPage]
class PDTestRoute extends _i5.PageRouteInfo<void> {
  const PDTestRoute({List<_i5.PageRouteInfo>? children})
    : super(PDTestRoute.name, initialChildren: children);

  static const String name = 'PDTestRoute';

  static _i5.PageInfo page = _i5.PageInfo(
    name,
    builder: (data) {
      return const _i1.PDTestPage();
    },
  );
}

/// generated route for
/// [_i2.PipelineDetailsPage]
class PipelineDetailsRoute extends _i5.PageRouteInfo<PipelineDetailsRouteArgs> {
  PipelineDetailsRoute({
    required String id,
    _i6.Key? key,
    List<_i5.PageRouteInfo>? children,
  }) : super(
         PipelineDetailsRoute.name,
         args: PipelineDetailsRouteArgs(id: id, key: key),
         rawPathParams: {'id': id},
         initialChildren: children,
       );

  static const String name = 'PipelineDetailsRoute';

  static _i5.PageInfo page = _i5.PageInfo(
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

  final _i6.Key? key;

  @override
  String toString() {
    return 'PipelineDetailsRouteArgs{id: $id, key: $key}';
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    if (other is! PipelineDetailsRouteArgs) return false;
    return id == other.id && key == other.key;
  }

  @override
  int get hashCode => id.hashCode ^ key.hashCode;
}

/// generated route for
/// [_i3.PipelinesPage]
class PipelinesRoute extends _i5.PageRouteInfo<void> {
  const PipelinesRoute({List<_i5.PageRouteInfo>? children})
    : super(PipelinesRoute.name, initialChildren: children);

  static const String name = 'PipelinesRoute';

  static _i5.PageInfo page = _i5.PageInfo(
    name,
    builder: (data) {
      return const _i3.PipelinesPage();
    },
  );
}

/// generated route for
/// [_i4.ResultsPage]
class ResultsRoute extends _i5.PageRouteInfo<void> {
  const ResultsRoute({List<_i5.PageRouteInfo>? children})
    : super(ResultsRoute.name, initialChildren: children);

  static const String name = 'ResultsRoute';

  static _i5.PageInfo page = _i5.PageInfo(
    name,
    builder: (data) {
      return const _i4.ResultsPage();
    },
  );
}
