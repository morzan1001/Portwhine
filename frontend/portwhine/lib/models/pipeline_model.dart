import 'dart:math';

import 'package:portwhine/api/api.dart' as gen;
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_definition.dart';
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
              y: (positionMap['y'] as num).toDouble(),
            )
          : NodePosition(x: 0, y: 0);

      // Extract inputs/outputs if available.
      // Backend has historically used both singular ('input'/'output') and plural ('inputs'/'outputs') keys.
      final inputsList =
          ((data['inputs'] ?? data['input']) as List?)?.cast<String>() ?? [];
      final outputsList =
          ((data['outputs'] ?? data['output']) as List?)?.cast<String>() ?? [];

      final inputs = {for (var i in inputsList) i: String};
      final outputs = {for (var i in outputsList) i: String};

      // Config is everything else
      final config = Map<String, dynamic>.from(data);
      config.remove('id');
      config.remove('status');
      config.remove('gridPosition');
      config.remove('input');
      config.remove('output');
      config.remove('inputs');
      config.remove('outputs');
      config.remove('instanceHealth');

      nodes.add(
        NodeModel(
          id: id,
          name: name,
          inputs: inputs,
          outputs: outputs,
          position: position,
          config: config,
        ),
      );
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
        // Backend edge (latest): { "source": "id", "target": "id", "source_port": "...", "target_port": "..." }
        // Some older persisted documents may still have {from,to}.
        // We need to find the nodes to get coordinates for LineModel
        final fromId = (e['source'] ?? e['from']) as String;
        final toId = (e['target'] ?? e['to']) as String;

        // Update inputNodes map on the target node
        final targetNodeIndex = nodes.indexWhere((n) => n.id == toId);
        if (targetNodeIndex != -1) {
          final targetNode = nodes[targetNodeIndex];
          // We need a key for the input.
          // If we don't have port info in edge, we generate one.
          final key = 'Input_${fromId.substring(0, min(4, fromId.length))}';

          nodes[targetNodeIndex] = targetNode.copyWith(
            inputNodes: {...targetNode.inputNodes, key: fromId},
          );
        }

        // Create LineModel for visualization
        final fromNode = nodes.firstWhere(
          (n) => n.id == fromId,
          orElse: () => NodeModel(),
        );
        final toNode = nodes.firstWhere(
          (n) => n.id == toId,
          orElse: () => NodeModel(),
        );

        if (fromNode.id.isNotEmpty && toNode.id.isNotEmpty) {
          edges.add(
            LineModel(
              startX: fromNode.position?.x ?? 0,
              startY: fromNode.position?.y ?? 0,
              endX: toNode.position?.x ?? 0,
              endY: toNode.position?.y ?? 0,
            ),
          );
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

    Map<String, dynamic> buildConfigDefaults(NodeModel node) {
      final definition = node.definition;
      if (definition == null) return {};

      final defaults = <String, dynamic>{};
      for (final field in definition.configFields) {
        if (field.defaultValue != null) {
          defaults[field.name] = field.defaultValue;
          continue;
        }

        if (field.required == true) {
          // Only set safe defaults for required fields where it's obvious.
          // In particular, CertstreamTrigger requires a regex.
          if (field.type == FieldType.regex) {
            defaults[field.name] = '.*';
          }
        }
      }

      return defaults;
    }

    for (var node in nodes) {
      final effectiveConfig = {...buildConfigDefaults(node), ...node.config};

      // Build node config with position
      final nodeConfig = {
        ...effectiveConfig,
        'id': node.id,
        'gridPosition': node.position != null
            ? {'x': node.position!.x, 'y': node.position!.y}
            : {'x': 0, 'y': 0},
      };

      if (node.name.endsWith('Trigger')) {
        trigger = {node.name: nodeConfig};
      } else if (node.name.endsWith('Worker')) {
        workers.add({node.name: nodeConfig});
      }

      node.inputNodes.forEach((inputName, sourceNodeId) {
        final sourceNode = nodes.firstWhere(
          (n) => n.id == sourceNodeId,
          orElse: () => NodeModel(),
        );

        String targetPortId = inputName;
        final targetPort = node.inputPorts
            .where((p) => p.id == targetPortId)
            .cast<PortDefinition?>()
            .firstWhere((p) => p != null, orElse: () => null);

        final targetDataType = targetPort?.dataType;

        String? sourcePortId;
        if (targetDataType != null) {
          final matchingSourcePort = sourceNode.outputPorts
              .where((p) => p.dataType == targetDataType)
              .cast<PortDefinition?>()
              .firstWhere((p) => p != null, orElse: () => null);
          sourcePortId = matchingSourcePort?.id;
        }

        sourcePortId ??= sourceNode.outputPorts.isNotEmpty
            ? sourceNode.outputPorts.first.id
            : null;

        edgesList.add({
          'source': sourceNodeId,
          'target': node.id,
          'source_port': sourcePortId,
          'target_port': targetPortId,
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

  /// Convert this model to a Pipeline for the API.
  gen.Pipeline toPipeline() {
    gen.TriggerConfig? triggerConfig;
    List<gen.WorkerConfig> workerConfigs = [];
    List<gen.Edge> edgesList = [];

    for (var node in nodes) {
      final gridPos = gen.GridPosition(
        x: node.position?.x ?? 0,
        y: node.position?.y ?? 0,
      );

      if (node.name.endsWith('Trigger')) {
        triggerConfig = gen.TriggerConfig(gridPosition: gridPos);
      } else if (node.name.endsWith('Worker')) {
        workerConfigs.add(
          gen.WorkerConfig(
            gridPosition: gridPos,
            numberOfInstances: node.config['numberOfInstances'] as int? ?? 1,
          ),
        );
      }

      // Create edges from inputNodes
      node.inputNodes.forEach((_, sourceNodeId) {
        edgesList.add(gen.Edge(source: sourceNodeId, target: node.id));
      });
    }

    return gen.Pipeline(
      name: name,
      trigger: triggerConfig,
      worker: workerConfigs,
      edges: edgesList,
    );
  }

  /// Convert this model to a PipelinePatch for the API.
  gen.PipelinePatch toPipelinePatch() {
    final payload = toMap();
    return gen.PipelinePatch(
      id: id,
      name: name,
      trigger: payload['trigger'],
      worker: payload['worker'],
      edges: payload['edges'],
    );
  }
}
