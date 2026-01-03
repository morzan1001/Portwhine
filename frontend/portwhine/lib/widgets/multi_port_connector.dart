import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/global/theme.dart';
import 'package:portwhine/models/node_definition.dart';
import 'package:portwhine/models/node_model.dart';

/// Widget for displaying multiple output ports on a node.
class MultiPortOutputs extends StatelessWidget {
  const MultiPortOutputs({required this.model, required this.ports, super.key});

  final NodeModel model;
  final List<PortDefinition> ports;

  @override
  Widget build(BuildContext context) {
    if (ports.isEmpty) {
      return const SizedBox.shrink();
    }

    // Calculate spacing based on node height and port count
    final totalHeight = nodeHeight - 32; // Account for padding
    // Port spacing is handled by spaceEvenly in Column
    final _ = ports.length > 1
        ? totalHeight / (ports.length + 1)
        : totalHeight / 2;

    return SizedBox(
      height: nodeHeight,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: ports.asMap().entries.map((entry) {
          final index = entry.key;
          final port = entry.value;
          return _PortOutput(model: model, port: port, portIndex: index);
        }).toList(),
      ),
    );
  }
}

/// Widget for displaying multiple input ports on a node.
class MultiPortInputs extends StatelessWidget {
  const MultiPortInputs({required this.model, required this.ports, super.key});

  final NodeModel model;
  final List<PortDefinition> ports;

  @override
  Widget build(BuildContext context) {
    if (ports.isEmpty) {
      return const SizedBox.shrink();
    }

    return SizedBox(
      height: nodeHeight,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: ports.asMap().entries.map((entry) {
          final index = entry.key;
          final port = entry.value;
          return _PortInput(model: model, port: port, portIndex: index);
        }).toList(),
      ),
    );
  }
}

/// Single output port connector with drag capability.
class _PortOutput extends StatefulWidget {
  const _PortOutput({
    required this.model,
    required this.port,
    required this.portIndex,
  });

  final NodeModel model;
  final PortDefinition port;
  final int portIndex;

  @override
  State<_PortOutput> createState() => _PortOutputState();
}

class _PortOutputState extends State<_PortOutput> {
  bool _isDragging = false;
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final portColor = widget.port.dataType.color;
    final connectingLineCubit = context.read<ConnectingLineCubit>();
    final canvasCubit = context.read<CanvasCubit>();

    return MouseRegion(
      cursor: SystemMouseCursors.grab,
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: Listener(
        onPointerDown: (event) {
          // Start line at the output connector position of the node
          final nodePos = widget.model.position;
          if (nodePos == null) return;
          // Calculate Y offset for this port based on its index
          final portCount = widget.model.outputPorts.length;
          final spacing = nodeHeight / (portCount + 1);
          final portY = spacing * (widget.portIndex + 1);
          final startX = nodePos.x + nodeWidth;
          final startY = nodePos.y + portY;
          connectingLineCubit.init(startX, startY);
        },
        onPointerMove: (event) {
          if (connectingLineCubit.state == null) return;
          final canvas = canvasCubit.state;
          final canvasPos = canvas.globalToCanvas(event.position);
          connectingLineCubit.updateLine(canvasPos.dx, canvasPos.dy);
        },
        // Note: Don't remove line on pointerUp - let DragTarget.onAccept handle it
        onPointerCancel: (event) {
          connectingLineCubit.remove();
        },
        child: Tooltip(
          message: '${widget.port.label} (${widget.port.dataType.displayName})',
          child: Draggable<Map>(
            hitTestBehavior: HitTestBehavior.opaque,
            onDragStarted: () {
              if (mounted) setState(() => _isDragging = true);
            },
            onDragUpdate: (details) {
              // Line update is handled by Listener.onPointerMove
            },
            onDragEnd: (details) {
              if (mounted) setState(() => _isDragging = false);
              // Remove line after drag ends
              Future.delayed(const Duration(milliseconds: 50), () {
                connectingLineCubit.remove();
              });
            },
            onDraggableCanceled: (velocity, offset) {
              if (mounted) setState(() => _isDragging = false);
              connectingLineCubit.remove();
            },
            data: {
              'node': widget.model,
              'port': widget.port,
              'portIndex': widget.portIndex,
            },
            dragAnchorStrategy: pointerDragAnchorStrategy,
            feedback: Transform.translate(
              offset: const Offset(-12, -12),
              child: Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: portColor,
                  boxShadow: [
                    BoxShadow(
                      color: portColor.withValues(alpha: 0.5),
                      blurRadius: 12,
                      spreadRadius: 2,
                    ),
                  ],
                ),
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Port label
                if (_isHovered || _isDragging)
                  AnimatedOpacity(
                    duration: const Duration(milliseconds: 150),
                    opacity: _isHovered || _isDragging ? 1.0 : 0.0,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      margin: const EdgeInsets.only(right: 4),
                      decoration: BoxDecoration(
                        color: portColor.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(
                          color: portColor.withValues(alpha: 0.3),
                          width: 1,
                        ),
                      ),
                      child: Text(
                        widget.port.dataType.displayName,
                        style: style(
                          size: 9,
                          weight: FontWeight.w600,
                          color: portColor,
                        ),
                      ),
                    ),
                  ),
                // Connector dot
                AnimatedScale(
                  scale: _isDragging ? 1.2 : (_isHovered ? 1.1 : 1.0),
                  duration: const Duration(milliseconds: 150),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 150),
                    width: _isHovered ? 20 : 16,
                    height: _isHovered ? 20 : 16,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: colors.surface,
                      border: Border.all(
                        color: _isHovered || _isDragging
                            ? portColor
                            : colors.textTertiary,
                        width: 2,
                      ),
                      boxShadow: _isHovered || _isDragging
                          ? [
                              BoxShadow(
                                color: portColor.withValues(alpha: 0.5),
                                blurRadius: 8,
                                spreadRadius: 1,
                              ),
                            ]
                          : null,
                    ),
                    child: Center(
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 150),
                        width: _isHovered ? 8 : 6,
                        height: _isHovered ? 8 : 6,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: _isHovered || _isDragging
                              ? portColor
                              : colors.textTertiary,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Single input port connector with drop capability.
class _PortInput extends StatefulWidget {
  const _PortInput({
    required this.model,
    required this.port,
    required this.portIndex,
  });

  final NodeModel model;
  final PortDefinition port;
  final int portIndex;

  @override
  State<_PortInput> createState() => _PortInputState();
}

class _PortInputState extends State<_PortInput> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final portColor = widget.port.dataType.color;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: Tooltip(
        message: '${widget.port.label} (${widget.port.dataType.displayName})',
        child: DragTarget<Map>(
          builder: (context, candidateData, rejectedData) {
            final isReceiving = candidateData.isNotEmpty;

            // Check if the incoming data type matches
            bool isCompatible = true;
            if (isReceiving && candidateData.first != null) {
              final incomingPort =
                  candidateData.first!['port'] as PortDefinition?;
              if (incomingPort != null) {
                isCompatible = incomingPort.dataType == widget.port.dataType;
              }
            }

            return Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Connector dot
                AnimatedScale(
                  scale: isReceiving ? 1.3 : (_isHovered ? 1.1 : 1.0),
                  duration: const Duration(milliseconds: 150),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 150),
                    width: _isHovered || isReceiving ? 20 : 16,
                    height: _isHovered || isReceiving ? 20 : 16,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: isReceiving && !isCompatible
                          ? colors.error.withValues(alpha: 0.2)
                          : colors.surface,
                      border: Border.all(
                        color: isReceiving
                            ? (isCompatible ? portColor : colors.error)
                            : (_isHovered ? portColor : colors.textTertiary),
                        width: 2,
                      ),
                      boxShadow: _isHovered || isReceiving
                          ? [
                              BoxShadow(
                                color:
                                    (isReceiving && !isCompatible
                                            ? colors.error
                                            : portColor)
                                        .withValues(alpha: 0.5),
                                blurRadius: isReceiving ? 12 : 8,
                                spreadRadius: isReceiving ? 2 : 1,
                              ),
                            ]
                          : null,
                    ),
                    child: Center(
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 150),
                        width: _isHovered || isReceiving ? 8 : 6,
                        height: _isHovered || isReceiving ? 8 : 6,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: isReceiving
                              ? (isCompatible ? portColor : colors.error)
                              : (_isHovered ? portColor : colors.textTertiary),
                        ),
                      ),
                    ),
                  ),
                ),
                // Port label
                if (_isHovered || isReceiving)
                  AnimatedOpacity(
                    duration: const Duration(milliseconds: 150),
                    opacity: _isHovered || isReceiving ? 1.0 : 0.0,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      margin: const EdgeInsets.only(left: 4),
                      decoration: BoxDecoration(
                        color: portColor.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(
                          color: portColor.withValues(alpha: 0.3),
                          width: 1,
                        ),
                      ),
                      child: Text(
                        widget.port.dataType.displayName,
                        style: style(
                          size: 9,
                          weight: FontWeight.w600,
                          color: portColor,
                        ),
                      ),
                    ),
                  ),
              ],
            );
          },
          onWillAcceptWithDetails: (details) {
            // Check data type compatibility
            final incomingPort = details.data['port'] as PortDefinition?;
            if (incomingPort != null) {
              return incomingPort.dataType == widget.port.dataType;
            }
            return true; // Accept if no port info (legacy)
          },
          onAcceptWithDetails: (details) {
            final outputNode = details.data['node'] as NodeModel;
            final inputNode = widget.model;

            final outputPort = details.data['port'] as PortDefinition?;
            final inputPort = widget.port;

            BlocProvider.of<NodesCubit>(context).addConnection(
              outputNode,
              inputNode,
              outputPort: outputPort,
              inputPort: inputPort,
            );

            BlocProvider.of<LinesCubit>(
              context,
            ).updateLines(BlocProvider.of<NodesCubit>(context).state);
          },
        ),
      ),
    );
  }
}
