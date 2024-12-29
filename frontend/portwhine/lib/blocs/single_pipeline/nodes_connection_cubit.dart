import 'package:bloc/bloc.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';

class NodesCubit extends Cubit<List<NodeModel>> {
  NodesCubit()
      : super(
          [
            NodeModel(
              id: "Test",
              name: 'Domain',
              inputs: {'Input 1': int, 'Input 2': int},
              outputs: {'Output': int},
              position: NodePosition(x: 2550, y: 2100),
            ),
            // NodeModel(
            //   id: generateId(),
            //   name: 'Nmap',
            //   inputs: {'Input 1': int, 'Input 2': int},
            //   outputs: {'Output': int},
            //   position: NodePosition(x: 1700, y: 1150),
            // ),
            // NodeModel(
            //   id: generateId(),
            //   name: 'Nikto',
            //   inputs: {'Input 1': int, 'Input 2': int},
            //   outputs: {'Output': int},
            //   position: NodePosition(x: 1500, y: 1350),
            // ),
          ],
        );

  void addNode(NodeModel model) {
    emit([...state, model]);
  }

  void addConnection(NodeModel output, NodeModel input) {
    emit(
      state.map((node) {
        if (node.id == input.id) {
          final updatedInputNodes = {...node.inputNodes, 'Input 1': output.id};
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
