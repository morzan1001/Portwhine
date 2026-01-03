part of 'pipeline_status_cubit.dart';

/// Base state for pipeline status.
sealed class PipelineStatusState extends Equatable {
  const PipelineStatusState();

  @override
  List<Object?> get props => [];
}

/// Initial state - not connected.
final class PipelineStatusInitial extends PipelineStatusState {}

/// Connecting to WebSocket.
final class PipelineStatusConnecting extends PipelineStatusState {
  final String pipelineId;

  const PipelineStatusConnecting({required this.pipelineId});

  @override
  List<Object?> get props => [pipelineId];
}

/// Connected and receiving updates.
final class PipelineStatusConnected extends PipelineStatusState {
  final String pipelineId;
  final PipelineStatusUpdate status;

  const PipelineStatusConnected({
    required this.pipelineId,
    required this.status,
  });

  @override
  List<Object?> get props => [pipelineId, status];

  /// Get status for a specific node.
  NodeStatusInfo? getNodeStatus(String nodeId) {
    return status.getNodeStatus(nodeId);
  }

  /// Get the overall pipeline status.
  NodeStatus get pipelineStatus => status.pipelineStatus;

  /// Check if any node has an error.
  bool get hasErrors => status.nodes.any((n) => n.status.isError);

  /// Check if all nodes are completed.
  bool get isCompleted => status.nodes.every((n) => n.status.isCompleted);

  /// Check if the pipeline is running.
  bool get isRunning => status.pipelineStatus.isActive;
}

/// Error state.
final class PipelineStatusError extends PipelineStatusState {
  final String pipelineId;
  final String message;

  const PipelineStatusError({
    required this.pipelineId,
    required this.message,
  });

  @override
  List<Object?> get props => [pipelineId, message];
}
