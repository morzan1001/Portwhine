import 'package:portwhine/models/node_model.dart';

class PipelineModel {
  String name;
  String? status;
  List<NodeModel> nodes;
  int totalNodes, nodesCompleted;
  int runningTime, expectedTime;
  String? currentNode;
  int? currentRunningTime;
  int errors;
  bool completed;

  PipelineModel({
    this.name = 'Test Workflow',
    this.status,
    this.nodes = const [],
    this.totalNodes = 0,
    this.nodesCompleted = 0,
    this.runningTime = 0,
    this.expectedTime = 0,
    this.currentNode,
    this.currentRunningTime,
    this.errors = 0,
    this.completed = false,
  });

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'status': status,
      'nodes': nodes.map((node) => node.toMap()).toList(),
      'totalNodes': totalNodes,
      'nodesCompleted': nodesCompleted,
      'runningTime': runningTime,
      'expectedTime': expectedTime,
      'currentNode': currentNode,
      'currentRunningTime': currentRunningTime,
      'errors': errors,
      'completed': completed,
    };
  }

  static PipelineModel fromMap(Map<String, dynamic> map) {
    return PipelineModel(
      name: map['name'],
      status: map['status'],
      nodes: List<NodeModel>.from(
          map['nodes'].map((node) => NodeModel.fromMap(node))),
      totalNodes: map['totalNodes'],
      nodesCompleted: map['nodesCompleted'],
      runningTime: map['runningTime'],
      expectedTime: map['expectedTime'],
      currentNode: map['currentNode'],
      currentRunningTime: map['currentRunningTime'],
      errors: map['errors'],
      completed: map['completed'],
    );
  }

  PipelineModel copyWith({
    String? name,
    String? status,
    List<NodeModel>? nodes,
    int? totalNodes,
    int? nodesCompleted,
    int? runningTime,
    int? expectedTime,
    String? currentNode,
    int? currentRunningTime,
    int? errors,
    bool? completed,
  }) {
    return PipelineModel(
      name: name ?? this.name,
      status: status ?? this.status,
      nodes: nodes ?? this.nodes,
      totalNodes: totalNodes ?? this.totalNodes,
      nodesCompleted: nodesCompleted ?? this.nodesCompleted,
      runningTime: runningTime ?? this.runningTime,
      expectedTime: expectedTime ?? this.expectedTime,
      currentNode: currentNode ?? this.currentNode,
      currentRunningTime: currentRunningTime ?? this.currentRunningTime,
      errors: errors ?? this.errors,
      completed: completed ?? this.completed,
    );
  }
}
