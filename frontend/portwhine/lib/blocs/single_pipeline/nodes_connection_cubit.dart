import 'dart:math';

import 'package:bloc/bloc.dart';
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';

class NodesCubit extends Cubit<List<NodeModel>> {
  NodesCubit() : super([]);

  void setNodes(List<NodeModel> nodes) {
    emit(nodes);
  }

  void addNode(NodeModel model) {
    // Ensure the node has a unique ID if it doesn't have one
    final nodeToAdd =
        model.id.isEmpty ? model.copyWith(id: _generateId()) : model;

    emit([...state, nodeToAdd]);
  }

  void removeNode(String id) {
    emit(state.where((node) => node.id != id).toList());
  }

  void addConnection(NodeModel output, NodeModel input) {
    // Check for compatible types
    // The keys of inputs/outputs represent the types (e.g., "HTTP", "IP")
    final outputTypes = output.outputs.keys.toSet();
    final inputTypes = input.inputs.keys.toSet();

    final commonTypes = outputTypes.intersection(inputTypes);

    if (commonTypes.isEmpty) {
      // No compatible types found
      // TODO: Emit a state or callback to notify UI of failure?
      // For now, just return without connecting.
      return;
    }

    emit(
      state.map((node) {
        if (node.id == input.id) {
          // Avoid duplicate connections
          if (node.inputNodes.containsValue(output.id)) {
            return node;
          }
          // Use the common type as the key if possible, or generate unique key
          // Ideally we map specific output port to specific input port.
          // For now, we just use the first common type as the "port" name if available
          final portName = commonTypes.first;

          // Check if this port is already occupied?
          // If we allow multiple inputs per port, we might need a list.
          // But the current model is Map<String, String> (InputName -> SourceNodeId).
          // So one source per input port.

          // If the port is already taken, we can't connect.
          if (node.inputNodes.containsKey(portName)) {
            // Port occupied
            return node;
          }

          final updatedInputNodes = {...node.inputNodes, portName: output.id};
          return node.copyWith(inputNodes: updatedInputNodes);
        }
        return node;
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

  void updateNode(NodeModel updatedNode) {
    emit(
      state.map((node) {
        if (node.id == updatedNode.id) {
          return updatedNode;
        }
        return node;
      }).toList(),
    );
  }

  String _generateId() {
    return DateTime.now().millisecondsSinceEpoch.toString() +
        Random().nextInt(1000).toString();
  }
}

class LinesCubit extends Cubit<List<LineModel>> {
  LinesCubit() : super([]);

  void updateLines(List<NodeModel> nodes) {
    List<LineModel> lines = [];

    for (var endNode in nodes) {
      if (endNode.inputNodes.isNotEmpty) {
        for (var entry in endNode.inputNodes.entries) {
          var startNode = nodes.firstWhere(
            (n) => n.id == entry.value,
          );

          double startX = startNode.position!.x;
          double startY = startNode.position!.y;

          double endX = endNode.position!.x;
          double endY = endNode.position!.y;

          lines.add(
            LineModel(
              startX: startX,
              startY: startY,
              endX: endX,
              endY: endY,
            ),
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
    emit(
      LineModel(startX: x, startY: y, endX: x, endY: y),
    );
  }

  void updateLine(double x, double y) {
    emit(state!.copyWith(endX: x, endY: y));
  }

  void remove() {
    emit(null);
  }
}
