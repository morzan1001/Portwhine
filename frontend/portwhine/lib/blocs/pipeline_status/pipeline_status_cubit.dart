import 'dart:async';

import 'package:bloc/bloc.dart';
import 'package:equatable/equatable.dart';

import 'package:portwhine/api/websocket_service.dart';
import 'package:portwhine/models/node_status.dart';

part 'pipeline_status_state.dart';

/// Cubit for managing real-time pipeline status via WebSocket.
///
/// This cubit connects to the WebSocket service and emits status updates
/// as they are received from the backend.
class PipelineStatusCubit extends Cubit<PipelineStatusState> {
  PipelineStatusCubit() : super(PipelineStatusInitial());

  StreamSubscription<PipelineStatusUpdate>? _subscription;
  String? _currentPipelineId;

  /// Start listening to status updates for a pipeline.
  Future<void> connectToPipeline(String pipelineId) async {
    if (_currentPipelineId == pipelineId) return;

    await disconnect();
    _currentPipelineId = pipelineId;

    emit(PipelineStatusConnecting(pipelineId: pipelineId));

    try {
      await pipelineWebSocketService.connectToPipeline(pipelineId);

      _subscription = pipelineWebSocketService.statusStream?.listen(
        (update) {
          emit(PipelineStatusConnected(
            pipelineId: pipelineId,
            status: update,
          ));
        },
        onError: (error) {
          emit(PipelineStatusError(
            pipelineId: pipelineId,
            message: error.toString(),
          ));
        },
      );
    } catch (e) {
      emit(PipelineStatusError(
        pipelineId: pipelineId,
        message: e.toString(),
      ));
    }
  }

  /// Disconnect from the current pipeline.
  Future<void> disconnect() async {
    await _subscription?.cancel();
    _subscription = null;
    await pipelineWebSocketService.disconnect();
    _currentPipelineId = null;
    emit(PipelineStatusInitial());
  }

  /// Get the current status of a specific node.
  NodeStatusInfo? getNodeStatus(String nodeId) {
    final currentState = state;
    if (currentState is PipelineStatusConnected) {
      return currentState.status.getNodeStatus(nodeId);
    }
    return null;
  }

  @override
  Future<void> close() async {
    await disconnect();
    return super.close();
  }
}
