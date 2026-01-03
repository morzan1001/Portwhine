// Node status models for real-time pipeline status updates via WebSocket.

/// Possible status values for a node.
enum NodeStatus {
  pending,
  starting,
  running,
  paused,
  stopped,
  completed,
  restarting,
  oomKilled,
  dead,
  error,
  unknown;

  static NodeStatus fromString(String? value) {
    if (value == null) return NodeStatus.unknown;

    switch (value.toLowerCase()) {
      case 'pending':
        return NodeStatus.pending;
      case 'starting':
        return NodeStatus.starting;
      case 'running':
        return NodeStatus.running;
      case 'paused':
        return NodeStatus.paused;
      case 'stopped':
        return NodeStatus.stopped;
      case 'completed':
        return NodeStatus.completed;
      case 'restarting':
        return NodeStatus.restarting;
      case 'oomkilled':
      case 'oom_killed':
        return NodeStatus.oomKilled;
      case 'dead':
        return NodeStatus.dead;
      case 'error':
        return NodeStatus.error;
      default:
        return NodeStatus.unknown;
    }
  }

  String get displayName {
    switch (this) {
      case NodeStatus.pending:
        return 'Pending';
      case NodeStatus.starting:
        return 'Starting';
      case NodeStatus.running:
        return 'Running';
      case NodeStatus.paused:
        return 'Paused';
      case NodeStatus.stopped:
        return 'Stopped';
      case NodeStatus.completed:
        return 'Completed';
      case NodeStatus.restarting:
        return 'Restarting';
      case NodeStatus.oomKilled:
        return 'Out of Memory';
      case NodeStatus.dead:
        return 'Dead';
      case NodeStatus.error:
        return 'Error';
      case NodeStatus.unknown:
        return 'Unknown';
    }
  }

  bool get isActive =>
      this == NodeStatus.running ||
      this == NodeStatus.starting ||
      this == NodeStatus.restarting;

  bool get isError =>
      this == NodeStatus.error ||
      this == NodeStatus.oomKilled ||
      this == NodeStatus.dead;

  bool get isCompleted => this == NodeStatus.completed;

  bool get isStopped => this == NodeStatus.stopped || this == NodeStatus.paused;
}

/// Health information for a single instance of a node.
class InstanceHealth {
  final int number;
  final NodeStatus health;

  const InstanceHealth({
    required this.number,
    required this.health,
  });

  factory InstanceHealth.fromJson(Map<String, dynamic> json) {
    return InstanceHealth(
      number: json['number'] as int,
      health: NodeStatus.fromString(json['health'] as String?),
    );
  }

  Map<String, dynamic> toJson() => {
        'number': number,
        'health': health.name,
      };
}

/// Status information for a single node in the pipeline.
class NodeStatusInfo {
  final String id;
  final String type;
  final NodeStatus status;
  final bool isTrigger;
  final List<InstanceHealth>? instanceHealth;
  final int numberOfInstances;

  const NodeStatusInfo({
    required this.id,
    required this.type,
    required this.status,
    this.isTrigger = false,
    this.instanceHealth,
    this.numberOfInstances = 0,
  });

  factory NodeStatusInfo.fromJson(Map<String, dynamic> json) {
    return NodeStatusInfo(
      id: json['id'] as String,
      type: json['type'] as String,
      status: NodeStatus.fromString(json['status'] as String?),
      isTrigger: json['is_trigger'] as bool? ?? false,
      instanceHealth: (json['instance_health'] as List?)
          ?.map((e) => InstanceHealth.fromJson(e as Map<String, dynamic>))
          .toList(),
      numberOfInstances: json['number_of_instances'] as int? ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type,
        'status': status.name,
        'is_trigger': isTrigger,
        'instance_health': instanceHealth?.map((e) => e.toJson()).toList(),
        'number_of_instances': numberOfInstances,
      };

  /// Get the overall health based on instance health.
  NodeStatus get overallHealth {
    if (instanceHealth == null || instanceHealth!.isEmpty) {
      return status;
    }

    // If any instance has error, the overall is error
    if (instanceHealth!.any((h) => h.health.isError)) {
      return NodeStatus.error;
    }

    // If all instances are running, overall is running
    if (instanceHealth!.every((h) => h.health == NodeStatus.running)) {
      return NodeStatus.running;
    }

    // If all instances are stopped, overall is stopped
    if (instanceHealth!.every((h) => h.health.isStopped)) {
      return NodeStatus.stopped;
    }

    // Mixed state - return running if any are running
    if (instanceHealth!.any((h) => h.health.isActive)) {
      return NodeStatus.running;
    }

    return status;
  }
}

/// Complete status update for a pipeline from WebSocket.
class PipelineStatusUpdate {
  final String pipelineId;
  final NodeStatus pipelineStatus;
  final List<NodeStatusInfo> nodes;

  const PipelineStatusUpdate({
    required this.pipelineId,
    required this.pipelineStatus,
    required this.nodes,
  });

  factory PipelineStatusUpdate.fromJson(Map<String, dynamic> json) {
    return PipelineStatusUpdate(
      pipelineId: json['pipeline_id'] as String,
      pipelineStatus: NodeStatus.fromString(json['pipeline_status'] as String?),
      nodes: (json['nodes'] as List?)
              ?.map((e) => NodeStatusInfo.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() => {
        'pipeline_id': pipelineId,
        'pipeline_status': pipelineStatus.name,
        'nodes': nodes.map((e) => e.toJson()).toList(),
      };

  /// Get status info for a specific node by ID.
  NodeStatusInfo? getNodeStatus(String nodeId) {
    try {
      return nodes.firstWhere((n) => n.id == nodeId);
    } catch (_) {
      return null;
    }
  }
}

/// WebSocket message types.
enum WebSocketMessageType {
  statusUpdate,
  ping,
  pong,
  subscribe,
  unsubscribe,
  error;

  static WebSocketMessageType fromString(String value) {
    switch (value) {
      case 'status_update':
        return WebSocketMessageType.statusUpdate;
      case 'ping':
        return WebSocketMessageType.ping;
      case 'pong':
        return WebSocketMessageType.pong;
      case 'subscribe':
        return WebSocketMessageType.subscribe;
      case 'unsubscribe':
        return WebSocketMessageType.unsubscribe;
      case 'error':
        return WebSocketMessageType.error;
      default:
        return WebSocketMessageType.error;
    }
  }
}

/// Parsed WebSocket message.
class WebSocketMessage {
  final WebSocketMessageType type;
  final dynamic data;

  const WebSocketMessage({
    required this.type,
    this.data,
  });

  factory WebSocketMessage.fromJson(Map<String, dynamic> json) {
    return WebSocketMessage(
      type: WebSocketMessageType.fromString(json['type'] as String? ?? 'error'),
      data: json['data'],
    );
  }

  /// Parse as a status update if applicable.
  PipelineStatusUpdate? get asStatusUpdate {
    if (type == WebSocketMessageType.statusUpdate && data != null) {
      return PipelineStatusUpdate.fromJson(data as Map<String, dynamic>);
    }
    return null;
  }
}
