import 'package:flutter/material.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/line_model.dart';
import 'package:portwhine/models/node_position.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/colors.dart';

class NodeEditor extends StatefulWidget {
  const NodeEditor({Key? key}) : super(key: key);

  @override
  _NodeEditorState createState() => _NodeEditorState();
}

class _NodeEditorState extends State<NodeEditor> {
  CanvasModel _canvasModel = CanvasModel();
  List<NodeModel> _nodes = [];
  List<LineModel> _lines = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Node Editor'),
      ),
      body: Stack(
        children: [
          _buildCanvas(),
          _buildNodeMenu(),
        ],
      ),
    );
  }

  Widget _buildCanvas() {
    return GestureDetector(
      onPanUpdate: (details) {
        setState(() {
          _canvasModel = _canvasModel.copyWith(
            position: NodePosition(
              x: _canvasModel.position.x + details.delta.dx,
              y: _canvasModel.position.y + details.delta.dy,
            ),
          );
        });
      },
      child: CustomPaint(
        painter: _CanvasPainter(_nodes, _lines),
        child: Container(),
      ),
    );
  }

  Widget _buildNodeMenu() {
    return Positioned(
      right: 16,
      top: 16,
      child: Column(
        children: [
          Draggable<NodeModel>(
            data: NodeModel(name: 'Node 1'),
            feedback: _buildNodeWidget(NodeModel(name: 'Node 1')),
            child: _buildNodeWidget(NodeModel(name: 'Node 1')),
          ),
          const SizedBox(height: 16),
          Draggable<NodeModel>(
            data: NodeModel(name: 'Node 2'),
            feedback: _buildNodeWidget(NodeModel(name: 'Node 2')),
            child: _buildNodeWidget(NodeModel(name: 'Node 2')),
          ),
        ],
      ),
    );
  }

  Widget _buildNodeWidget(NodeModel node) {
    return Container(
      width: nodeWidth,
      height: nodeHeight,
      decoration: BoxDecoration(
        color: CustomColors.sec,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Center(
        child: Text(
          node.name,
          style: TextStyle(color: CustomColors.white),
        ),
      ),
    );
  }

  void _onNodeDropped(NodeModel node, Offset position) {
    setState(() {
      _nodes.add(node.copyWith(
        position: NodePosition(x: position.dx, y: position.dy),
      ));
    });
  }

  void _onNodeRemoved(NodeModel node) {
    setState(() {
      _nodes.remove(node);
    });
  }

  void _onNodeLinked(NodeModel startNode, NodeModel endNode) {
    setState(() {
      _lines.add(LineModel(
        startX: startNode.position!.x,
        startY: startNode.position!.y,
        endX: endNode.position!.x,
        endY: endNode.position!.y,
      ));
    });
  }
}

class _CanvasPainter extends CustomPainter {
  final List<NodeModel> nodes;
  final List<LineModel> lines;

  _CanvasPainter(this.nodes, this.lines);

  @override
  void paint(Canvas canvas, Size size) {
    for (var line in lines) {
      final paint = Paint()
        ..color = CustomColors.secDark
        ..strokeWidth = 2;
      canvas.drawLine(
        Offset(line.startX, line.startY),
        Offset(line.endX, line.endY),
        paint,
      );
    }

    for (var node in nodes) {
      final paint = Paint()
        ..color = CustomColors.sec
        ..style = PaintingStyle.fill;
      canvas.drawRect(
        Rect.fromLTWH(node.position!.x, node.position!.y, nodeWidth, nodeHeight),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) {
    return true;
  }
}
