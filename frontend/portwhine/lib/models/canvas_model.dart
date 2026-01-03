import 'package:flutter/cupertino.dart';
import 'package:portwhine/models/position.dart';

class CanvasModel {
  Position position;
  double zoom;
  TransformationController? controller;
  Animation<Matrix4>? positionAnimation;
  AnimationController? positionController;
  GlobalKey? canvasKey;

  CanvasModel({
    this.position = Position.zero,
    this.zoom = 1,
    this.controller,
    this.positionAnimation,
    this.positionController,
    this.canvasKey,
  });

  CanvasModel copyWith({
    Position? position,
    double? zoom,
    TransformationController? controller,
    Animation<Matrix4>? positionAnimation,
    AnimationController? positionController,
    GlobalKey? canvasKey,
  }) {
    return CanvasModel(
      position: position ?? this.position,
      zoom: zoom ?? this.zoom,
      controller: controller ?? this.controller,
      positionAnimation: positionAnimation ?? this.positionAnimation,
      positionController: positionController ?? this.positionController,
      canvasKey: canvasKey ?? this.canvasKey,
    );
  }

  /// Get the current zoom level from the controller
  double get currentZoom {
    if (controller == null) return zoom;
    return controller!.value.getMaxScaleOnAxis();
  }

  /// Convert global screen position to canvas-local coordinates
  /// Converts global (screen) coordinates into the canvas "scene" coordinates
  /// of the InteractiveViewer.
  ///
  /// This should be used for anything that must line up with the mouse
  /// regardless of pan/zoom (node dragging, connection dragging, node drop).
  Offset globalToCanvas(Offset globalPosition, {RenderBox? viewerRenderBox}) {
    final transformationController = controller;
    if (transformationController == null) return globalPosition;

    final renderBox = viewerRenderBox ??
        canvasKey?.currentContext?.findRenderObject() as RenderBox?;
    if (renderBox == null) return globalPosition;

    final viewportPoint = renderBox.globalToLocal(globalPosition);
    return transformationController.toScene(viewportPoint);
  }
}
