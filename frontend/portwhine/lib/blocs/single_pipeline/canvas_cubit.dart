import 'package:bloc/bloc.dart';
import 'package:flutter/cupertino.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/models/position.dart';

class CanvasCubit extends Cubit<CanvasModel> {
  CanvasCubit() : super(CanvasModel());

  void setPosition(double x, double y) {
    emit(state.copyWith(position: Position(x, y)));
  }

  void setZoom(double zoom) {
    emit(state.copyWith(zoom: zoom));
  }

  void setController(TickerProvider tickerProvider) {
    final controller = TransformationController();
    controller.value.translate(-2000, -2000);
    final positionController = AnimationController(
      vsync: tickerProvider,
      duration: const Duration(milliseconds: 500),
    );
    positionController.addListener(() {
      state.controller!.value = state.positionAnimation!.value;
    });
    emit(
      state.copyWith(
        controller: controller,
        positionController: positionController,
        position: Position(
          controller.value.getTranslation().x,
          controller.value.getTranslation().y,
        ),
      ),
    );
  }

  void zoom(bool zoom) {
    if (state.controller == null) return;

    final currentZoom = state.controller!.value.getMaxScaleOnAxis();

    state.controller!.value.scale(
      (currentZoom + (zoom ? 0.1 : -0.1)) / currentZoom,
    );
    emit(
      state.copyWith(
        controller: state.controller,
        zoom: state.controller!.value.getMaxScaleOnAxis(),
      ),
    );
  }

  void changePosition(Position position) {
    final Matrix4 currentMatrix = state.controller!.value;
    final Matrix4 newMatrix = currentMatrix.clone()
      ..translate(
        position.x,
        position.y,
      );

    final animatedMatrix = Matrix4Tween(
      begin: currentMatrix,
      end: newMatrix,
    ).animate(
      CurvedAnimation(
        parent: state.positionController!,
        curve: Curves.easeInOut,
      ),
    );

    animatedMatrix.addListener(
      () {
        final position = animatedMatrix.value.getTranslation();
        emit(
          state.copyWith(
            positionAnimation: animatedMatrix,
            position: Position(position.x, position.y),
          ),
        );
      },
    );

    state.positionAnimation = animatedMatrix;
    state.positionController!.reset();
    state.positionController!.forward();
  }
}
