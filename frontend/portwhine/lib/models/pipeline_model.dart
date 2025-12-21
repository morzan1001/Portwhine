import 'dart:math';

import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';

class PipelineModel {
  final String id;
  final String name;
  final String status;
  final List<NodeModel> nodes;
  final List<LineModel> edges;

  PipelineModel({
    this.id = '',
    this.name = '',
    this.status = '',
    this.nodes = const [],
    this.edges = const [],
  });

  factory PipelineModel.fromMap(Map<String, dynamic> map) {
    List<NodeModel> nodes = [];
    List<LineModel> edges = [];

    // Helper to parse a node from the backend structure
    // Backend structure: { "WorkerName": { "id": "...", "gridPosition": {"x": 0, "y": 0}, ...config } }
    void parseNode(Map<String, dynamic> nodeMap) {
      if (nodeMap.isEmpty) return;
      final name = nodeMap.keys.first;
      final data = nodeMap[name] as Map<String, dynamic>;

      final id = data['id'] as String? ?? '';
      final positionMap = data['gridPosition'] as Map<String, dynamic>?;
      final position = positionMap != null
          ? NodePosition(
              x: (positionMap['x'] as num).toDouble(),
              y: (positionMap['y'] as num).toDouble())
          : NodePosition(x: 0, y: 0);

      // Extract inputs/outputs if available, otherwise empty (will be populated by repo if needed, or we just trust backend)
      // Backend sends 'input' and 'output' lists of strings/enums.
      final inputsList = (data['input'] as List?)?.cast<String>() ?? [];
      final outputsList = (data['output'] as List?)?.cast<String>() ?? [];

      final inputs = {for (var i in inputsList) i: String};
      final outputs = {for (var i in outputsList) i: String};

      // Config is everything else
      final config = Map<String, dynamic>.from(data);
      config.remove('id');
      config.remove('status');
      config.remove('gridPosition');
      config.remove('input');
      config.remove('output');
      config.remove('instanceHealth');

      nodes.add(NodeModel(
        id: id,
        name: name,
        inputs: inputs,
        outputs: outputs,
        position: position,
        config: config,
      ));
    }

    if (map['trigger'] != null) {
      parseNode(map['trigger'] as Map<String, dynamic>);
    }

    if (map['worker'] != null) {
      for (var w in map['worker'] as List) {
        parseNode(w as Map<String, dynamic>);
      }
    }

    if (map['edges'] != null) {
      for (var e in map['edges'] as List) {
        // Backend edge: { "from": "id", "to": "id" }
        // We need to find the nodes to get coordinates for LineModel
        final fromId = e['from'] as String;
        final toId = e['to'] as String;

        // Update inputNodes map on the target node
        final targetNodeIndex = nodes.indexWhere((n) => n.id == toId);
        if (targetNodeIndex != -1) {
          final targetNode = nodes[targetNodeIndex];
          // We need a key for the input.
          // If we don't have port info in edge, we generate one.
          final key = 'Input_${fromId.substring(0, min(4, fromId.length))}';

          nodes[targetNodeIndex] = targetNode
              .copyWith(inputNodes: {...targetNode.inputNodes, key: fromId});
        }

        // Create LineModel for visualization
        final fromNode =
            nodes.firstWhere((n) => n.id == fromId, orElse: () => NodeModel());
        final toNode =
            nodes.firstWhere((n) => n.id == toId, orElse: () => NodeModel());

        if (fromNode.id.isNotEmpty && toNode.id.isNotEmpty) {
          edges.add(LineModel(
            startX: fromNode.position?.x ?? 0,
            startY: fromNode.position?.y ?? 0,
            endX: toNode.position?.x ?? 0,
            endY: toNode.position?.y ?? 0,
          ));
        }
      }
    }

    return PipelineModel(
      id: map['id'] as String,
      name: map['name'] as String,
      status: map['status'] as String,
      nodes: nodes,
      edges: edges,
    );
  }

  Map<String, dynamic> toMap() {
    Map<String, dynamic>? trigger;
    List<Map<String, dynamic>> workers = [];
    List<Map<String, dynamic>> edgesList = [];

    for (var node in nodes) {
      if (node.name.endsWith('Trigger')) {
        trigger = {node.name: node.config};
      } else if (node.name.endsWith('Worker')) {
        workers.add({node.name: node.config});
      }

      node.inputNodes.forEach((inputName, sourceNodeId) {
        edgesList.add({
          'from': sourceNodeId,
          'to': node.id,
        });
      });
    }

    return {
      'id': id,
      'name': name,
      'status': status,
      if (trigger != null) 'trigger': trigger,
      'worker': workers,
      'edges': edgesList,
    };
  }

  PipelineModel copyWith({
    String? id,
    String? name,
    String? status,
    List<NodeModel>? nodes,
    List<LineModel>? edges,
  }) {
    return PipelineModel(
      id: id ?? this.id,
      name: name ?? this.name,
      status: status ?? this.status,
      nodes: nodes ?? this.nodes,
      edges: edges ?? this.edges,
    );
  }
}
