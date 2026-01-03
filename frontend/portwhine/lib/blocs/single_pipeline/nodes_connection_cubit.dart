import 'dart:math';

import 'package:bloc/bloc.dart';
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_definition.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';

class NodesCubit extends Cubit<List<NodeModel>> {
  NodesCubit() : super([]);

  void setNodes(List<NodeModel> nodes) {
    emit(nodes);
  }

  void addNode(NodeModel model) {
    // Ensure the node has a unique ID if it doesn't have one
    final nodeToAdd = model.id.isEmpty
        ? model.copyWith(id: _generateId())
        : model;

    emit([...state, nodeToAdd]);
  }

  void removeNode(String id) {
    emit(state.where((node) => node.id != id).toList());
  }

  void updateNode(NodeModel node) {
    emit(state.map((n) => n.id == node.id ? node : n).toList());
  }

  void clearErrors() {
    emit(state.map((n) => n.copyWith(error: '')).toList());
  }

  void addConnection(
    NodeModel output,
    NodeModel input, {
    PortDefinition? outputPort,
    PortDefinition? inputPort,
  }) {
    // Preferred path: use port definitions and data types.
    final outputPorts = output.outputPorts;
    final inputPorts = input.inputPorts;

    PortDefinition? resolvedOutputPort = outputPort;
    if (resolvedOutputPort == null && outputPorts.isNotEmpty) {
      resolvedOutputPort = outputPorts.first;
    }

    PortDefinition? resolvedInputPort = inputPort;
    if (resolvedInputPort == null && resolvedOutputPort != null) {
      for (final candidate in inputPorts) {
        if (candidate.dataType == resolvedOutputPort.dataType) {
          resolvedInputPort = candidate;
          break;
        }
      }
    }
    resolvedInputPort ??= inputPorts.isNotEmpty ? inputPorts.first : null;

    // If we have a resolved input port, use its ID as the key in inputNodes.
    if (resolvedInputPort != null) {
      final portKey = resolvedInputPort.id;

      emit(
        state.map((node) {
          if (node.id != input.id) return node;

          // Avoid duplicate connections from same output node.
          if (node.inputNodes.containsValue(output.id)) {
            return node;
          }

          // Current model allows one source per input port.
          if (node.inputNodes.containsKey(portKey)) {
            return node;
          }

          final updatedInputNodes = {...node.inputNodes, portKey: output.id};
          return node.copyWith(inputNodes: updatedInputNodes);
        }).toList(),
      );
      return;
    }

    // Legacy fallback: match by keys (older nodes used type strings as keys).
    final outputTypes = output.outputs.keys.toSet();
    final inputTypes = input.inputs.keys.toSet();
    final commonTypes = outputTypes.intersection(inputTypes);
    if (commonTypes.isEmpty) return;

    emit(
      state.map((node) {
        if (node.id != input.id) return node;

        if (node.inputNodes.containsValue(output.id)) {
          return node;
        }

        final portName = commonTypes.first;
        if (node.inputNodes.containsKey(portName)) {
          return node;
        }

        final updatedInputNodes = {...node.inputNodes, portName: output.id};
        return node.copyWith(inputNodes: updatedInputNodes);
      }).toList(),
    );
  }

  void moveNode(String id, NodePosition position) {
    emit(
      state.map((node) {
        if (node.id == id) {
          return node.copyWith(position: position);
        }
        return node;
      }).toList(),
    );
  }

  String _generateId() {
    final random = Random();

    String hex(int length) {
      return List.generate(
        length,
        (_) => random.nextInt(16).toRadixString(16),
      ).join();
    }

    final uuid =
        '${hex(8)}-${hex(4)}-4${hex(3)}-${(8 + random.nextInt(4)).toRadixString(16)}${hex(3)}-${hex(12)}';
    return 'new-$uuid';
  }
}

class LinesCubit extends Cubit<List<LineModel>> {
  LinesCubit() : super([]);

  void updateLines(List<NodeModel> nodes) {
    List<LineModel> lines = [];

    for (var endNode in nodes) {
      if (endNode.inputNodes.isNotEmpty) {
        for (var entry in endNode.inputNodes.entries) {
          var startNode = nodes.firstWhere((n) => n.id == entry.value);

          double startX = startNode.position!.x;
          double startY = startNode.position!.y;

          double endX = endNode.position!.x;
          double endY = endNode.position!.y;

          lines.add(
            LineModel(startX: startX, startY: startY, endX: endX, endY: endY),
          );
        }
      }
    }

    emit(lines);
  }
}

class ConnectingLineCubit extends Cubit<LineModel?> {
  ConnectingLineCubit() : super(null);

  void init(double x, double y) {
    emit(LineModel(startX: x, startY: y, endX: x, endY: y));
  }

  void updateLine(double x, double y) {
    // Only update if we have an active connecting line
    if (state == null) return;
    emit(state!.copyWith(endX: x, endY: y));
  }

  void remove() {
    emit(null);
  }
}
