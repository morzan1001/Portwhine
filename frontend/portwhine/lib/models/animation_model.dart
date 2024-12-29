import 'package:flutter/material.dart';

class AnimationModel<T> {
  AnimationController? controller;
  Animation<T>? animation;

  AnimationModel({
    this.controller,
    this.animation,
  });

  AnimationModel<T> copyWith({
    AnimationController? controller,
    Animation<T>? animation,
  }) {
    return AnimationModel(
      controller: controller ?? this.controller,
      animation: animation ?? this.animation,
    );
  }
}
