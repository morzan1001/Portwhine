import 'package:flutter/cupertino.dart';
import 'package:portwhine/models/position.dart';

class CanvasModel {
  Position position;
  double zoom;
  TransformationController? controller;
  Animation<Matrix4>? positionAnimation;
  AnimationController? positionController;

  CanvasModel({
    this.position = Position.zero,
    this.zoom = 1,
    this.controller,
    this.positionAnimation,
    this.positionController,
  });

  CanvasModel copyWith({
    Position? position,
    double? zoom,
    TransformationController? controller,
    Animation<Matrix4>? positionAnimation,
    AnimationController? positionController,
  }) {
    return CanvasModel(
      position: position ?? this.position,
      zoom: zoom ?? this.zoom,
      controller: controller ?? this.controller,
      positionAnimation: positionAnimation ?? this.positionAnimation,
      positionController: positionController ?? this.positionController,
    );
  }
}
