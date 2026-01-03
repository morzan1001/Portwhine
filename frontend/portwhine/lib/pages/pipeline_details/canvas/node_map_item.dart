import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/global/helpers.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/global/theme.dart';
import 'package:portwhine/models/node_definition.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/models/node_position.dart';
import 'package:portwhine/models/node_status.dart';
import 'package:portwhine/models/position.dart';
import 'package:portwhine/pages/pipeline_details/node_details.dart';
import 'package:portwhine/widgets/multi_port_connector.dart';
import 'package:portwhine/widgets/node_status_widgets.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeMapItem extends StatefulWidget {
  const NodeMapItem(this.model, {this.isPreview = false, super.key});

  final NodeModel model;
  final bool isPreview;

  @override
  State<NodeMapItem> createState() => _NodeMapItemState();
}

class _NodeMapItemState extends State<NodeMapItem> {
  bool _isHovered = false;
  Offset? _dragStartScene;
  Offset? _dragStartNode;

  void _onHoverChanged(bool hovered) {
    if (_isHovered != hovered) {
      setState(() => _isHovered = hovered);
    }
  }

  @override
  Widget build(BuildContext context) {
    return RepaintBoundary(
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          // Main node with GestureDetector
          MouseRegion(
            onEnter: (_) => _onHoverChanged(true),
            onExit: (_) => _onHoverChanged(false),
            child: GestureDetector(
              onTap: widget.isPreview
                  ? null
                  : () async {
                      final canvas = BlocProvider.of<CanvasCubit>(
                        context,
                      ).state;

                      final currentX = widget.model.position!.x;
                      final desiredX =
                          (((width(context) - nodeWidth) / 2) -
                              canvas.position.x) /
                          canvas.zoom;
                      final translateX = (desiredX - currentX).roundToDouble();

                      final currentY = widget.model.position!.y;
                      final desiredY = (100 - canvas.position.y) / canvas.zoom;
                      final translateY = (desiredY - currentY).roundToDouble();

                      BlocProvider.of<CanvasCubit>(
                        context,
                      ).changePosition(Position(translateX, translateY));

                      if (translateX != 0 || translateY != 0) {
                        await Future.delayed(const Duration(milliseconds: 500));
                      }

                      BlocProvider.of<SelectedNodeCubit>(
                        context,
                      ).setNode(widget.model);
                      await showNodeDetailsDialog(context, widget.model);
                      BlocProvider.of<SelectedNodeCubit>(context).removeNode();
                    },
              onPanStart: widget.isPreview
                  ? null
                  : (details) {
                      final canvas = context.read<CanvasCubit>().state;
                      final position = widget.model.position;
                      if (position == null) return;

                      _dragStartScene = canvas.globalToCanvas(
                        details.globalPosition,
                      );
                      _dragStartNode = Offset(position.x, position.y);
                    },
              onPanUpdate: widget.isPreview
                  ? null
                  : (details) {
                      final startScene = _dragStartScene;
                      final startNode = _dragStartNode;
                      if (startScene == null || startNode == null) return;

                      final canvas = context.read<CanvasCubit>().state;
                      final currentScene = canvas.globalToCanvas(
                        details.globalPosition,
                      );
                      final delta = currentScene - startScene;

                      final newX = startNode.dx + delta.dx;
                      final newY = startNode.dy + delta.dy;

                      context.read<NodesCubit>().moveNode(
                        widget.model.id,
                        NodePosition(x: newX, y: newY),
                      );

                      context.read<LinesCubit>().updateLines(
                        context.read<NodesCubit>().state,
                      );
                    },
              onPanEnd: widget.isPreview
                  ? null
                  : (_) {
                      _dragStartScene = null;
                      _dragStartNode = null;
                    },
              child: NodeActivityBorder(
                status: widget.model.status,
                isHovered: _isHovered,
                child: _buildNodeContainer(context),
              ),
            ),
          ),
          // Output connector(s) - outside GestureDetector for independent hit testing
          Positioned(
            right: -10,
            top: 0,
            bottom: 0,
            child: Center(
              child: _hasMultipleOutputs
                  ? MultiPortOutputs(
                      model: widget.model,
                      ports: widget.model.outputPorts,
                    )
                  : NodeOutput(widget.model),
            ),
          ),
          // Input connector(s) - outside GestureDetector for independent hit testing
          if (!widget.model.isTrigger)
            Positioned(
              left: -10,
              top: 0,
              bottom: 0,
              child: Center(
                child: _hasMultipleInputs
                    ? MultiPortInputs(
                        model: widget.model,
                        ports: widget.model.inputPorts,
                      )
                    : NodeInput(widget.model),
              ),
            ),
        ],
      ),
    );
  }

  bool get _hasMultipleOutputs => widget.model.outputPorts.length > 1;
  bool get _hasMultipleInputs => widget.model.inputPorts.length > 1;

  Widget _buildNodeContainer(BuildContext context) {
    final colors = context.colors;
    final nodeColor = widget.model.definition?.color ?? _getNodeTypeColor();
    final hasStatus = widget.model.statusInfo != null;
    // Defensive check for error being null (despite type definition)
    // ignore: unnecessary_null_comparison
    final hasError =
        widget.model.error != null && widget.model.error.isNotEmpty;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      width: nodeWidth,
      height: nodeHeight,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        color: colors.surface,
        border: Border.all(
          color: hasError
              ? colors.error
              : _isHovered
              ? nodeColor
              : widget.model.status.isError
              ? colors.error
              : Colors.transparent,
          width: hasError ? 3 : 2,
        ),
        boxShadow: [
          BoxShadow(
            color: hasError
                ? colors.error.withValues(alpha: 0.2)
                : _isHovered
                ? nodeColor.withValues(alpha: 0.3)
                : colors.shadow.withValues(alpha: 0.08),
            blurRadius: _isHovered || hasError ? 20 : 12,
            spreadRadius: _isHovered || hasError ? 2 : 0,
            offset: Offset(0, _isHovered ? 8 : 4),
          ),
          if (_isHovered)
            BoxShadow(
              color: nodeColor.withValues(alpha: 0.1),
              blurRadius: 40,
              spreadRadius: 8,
            ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header row with status and name
          Row(
            children: [
              // Status indicator (live or type color)
              if (hasStatus)
                NodeStatusIndicator(status: widget.model.status, size: 10)
              else
                Container(
                  width: 10,
                  height: 10,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: hasError ? colors.error : nodeColor,
                    boxShadow: [
                      BoxShadow(
                        color: (hasError ? colors.error : nodeColor).withValues(
                          alpha: 0.5,
                        ),
                        blurRadius: 6,
                        spreadRadius: 1,
                      ),
                    ],
                  ),
                ),
              const HorizontalSpacer(8),
              Expanded(
                child: Text(
                  // ignore: unnecessary_null_comparison
                  widget.model.definition?.name ?? widget.model.name ?? '',
                  style: style(
                    color: colors.textPrimary,
                    weight: FontWeight.w600,
                    size: 14,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              // Status badge (compact)
              if (hasStatus && widget.model.status != NodeStatus.unknown)
                NodeStatusBadge(status: widget.model.status, compact: true),
              if (hasError)
                Tooltip(
                  // ignore: unnecessary_null_comparison
                  message: widget.model.error ?? '',
                  child: Icon(
                    Icons.error_outline_rounded,
                    color: colors.error,
                    size: 18,
                  ),
                ),
            ],
          ),
          const VerticalSpacer(8),
          // Description (if available from definition)
          if (widget.model.description.isNotEmpty)
            Text(
              widget.model.description,
              style: style(size: 10, color: colors.textSecondary),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          const VerticalSpacer(8),
          // Ports visualization
          Expanded(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Input ports
                if (!widget.model.isTrigger)
                  Expanded(
                    child: _buildPortsList(
                      widget.model.inputPorts,
                      isInput: true,
                    ),
                  ),
                if (!widget.model.isTrigger &&
                    widget.model.outputPorts.isNotEmpty)
                  const HorizontalSpacer(8),
                // Output ports
                Expanded(
                  child: _buildPortsList(
                    widget.model.outputPorts,
                    isInput: false,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPortsList(List<PortDefinition> ports, {required bool isInput}) {
    if (ports.isEmpty) {
      // Fallback to legacy inputs/outputs
      final entries = isInput
          ? widget.model.inputs.entries.toList()
          : widget.model.outputs.entries.toList();
      return _buildIOList(entries, isInput: isInput);
    }

    return ListView.separated(
      separatorBuilder: (context, index) => const VerticalSpacer(4),
      itemCount: ports.length,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemBuilder: (context, i) {
        return _PortLabel(port: ports[i], isInput: isInput);
      },
    );
  }

  Widget _buildIOList(List<MapEntry> entries, {required bool isInput}) {
    if (entries.isEmpty) {
      return const SizedBox.shrink();
    }

    return ListView.separated(
      separatorBuilder: (context, index) => const VerticalSpacer(6),
      itemCount: entries.length,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemBuilder: (context, i) {
        return InputOutputItem(entries[i], isInput: isInput);
      },
    );
  }

  Color _getNodeTypeColor() => NodeHelper.getNodeColor(widget.model.name);
}

class ConnectorWidget extends StatefulWidget {
  const ConnectorWidget({
    this.isInput = false,
    this.isHovered = false,
    super.key,
  });

  final bool isInput;
  final bool isHovered;

  @override
  State<ConnectorWidget> createState() => _ConnectorWidgetState();
}

class _ConnectorWidgetState extends State<ConnectorWidget> {
  bool _isLocalHovered = false;

  @override
  Widget build(BuildContext context) {
    final color = widget.isInput
        ? NodeHelper.inputColor
        : NodeHelper.outputColor;

    final isActive = _isLocalHovered || widget.isHovered;

    return MouseRegion(
      onEnter: (_) => setState(() => _isLocalHovered = true),
      onExit: (_) => setState(() => _isLocalHovered = false),
      child: SizedBox(
        height: nodeHeight,
        child: Center(
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            width: isActive ? 22 : 18,
            height: isActive ? 22 : 18,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: MyColors.white,
              border: Border.all(
                color: isActive ? color : MyColors.darkGrey,
                width: 2,
              ),
              boxShadow: isActive
                  ? [
                      BoxShadow(
                        color: color.withValues(alpha: 0.5),
                        blurRadius: 12,
                        spreadRadius: 2,
                      ),
                    ]
                  : [],
            ),
            child: Center(
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 150),
                width: isActive ? 10 : 8,
                height: isActive ? 10 : 8,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isActive ? color : MyColors.darkGrey,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class NodeInput extends StatefulWidget {
  const NodeInput(this.model, {super.key});

  final NodeModel model;

  @override
  State<NodeInput> createState() => _NodeInputState();
}

class _NodeInputState extends State<NodeInput> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: DragTarget<Map>(
        builder: (context, candidateData, rejectedData) {
          final isReceiving = candidateData.isNotEmpty;
          return AnimatedScale(
            scale: isReceiving ? 1.3 : 1.0,
            duration: const Duration(milliseconds: 150),
            child: ConnectorWidget(
              isInput: true,
              isHovered: _isHovered || isReceiving,
            ),
          );
        },
        onWillAcceptWithDetails: (details) {
          // Accept if it's a different node (can't connect to self)
          final outputNode = details.data['node'] as NodeModel?;
          if (outputNode == null) return false;
          if (outputNode.id == widget.model.id) return false;

          // If we have port metadata, only accept compatible types.
          final outputPort = details.data['port'] as PortDefinition?;
          final inputPorts = widget.model.inputPorts;
          if (outputPort != null && inputPorts.isNotEmpty) {
            return inputPorts.any((p) => p.dataType == outputPort.dataType);
          }

          return true;
        },
        onAcceptWithDetails: (details) {
          final outputNode = details.data['node'] as NodeModel;
          final inputNode = widget.model;

          final outputPort = details.data['port'] as PortDefinition?;
          PortDefinition? inputPort;
          if (outputPort != null) {
            for (final candidate in inputNode.inputPorts) {
              if (candidate.dataType == outputPort.dataType) {
                inputPort = candidate;
                break;
              }
            }
          }

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
    );
  }
}

class NodeOutput extends StatefulWidget {
  const NodeOutput(this.model, {super.key});

  final NodeModel model;

  @override
  State<NodeOutput> createState() => _NodeOutputState();
}

class _NodeOutputState extends State<NodeOutput> {
  bool _isDragging = false;

  @override
  Widget build(BuildContext context) {
    final connectingLineCubit = context.read<ConnectingLineCubit>();
    final canvasCubit = context.read<CanvasCubit>();

    return MouseRegion(
      cursor: SystemMouseCursors.grab,
      child: Listener(
        onPointerDown: (event) {
          // Start line at the output connector position of the node
          final nodePos = widget.model.position;
          if (nodePos == null) return;
          final startX = nodePos.x + nodeWidth;
          final startY = nodePos.y + nodeHeight / 2;
          connectingLineCubit.init(startX, startY);
        },
        onPointerMove: (event) {
          if (connectingLineCubit.state == null) return;
          final canvas = canvasCubit.state;
          final canvasPos = canvas.globalToCanvas(event.position);
          connectingLineCubit.updateLine(canvasPos.dx, canvasPos.dy);
        },
        // Note: Don't remove line on pointerUp - let DragTarget.onAccept handle it
        // The line will be removed by onDraggableCanceled if not accepted
        onPointerCancel: (event) {
          connectingLineCubit.remove();
        },
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
            // Remove line after drag ends (whether accepted or not)
            // Small delay to allow DragTarget to process first
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
            'port': widget.model.outputPorts.isNotEmpty
                ? widget.model.outputPorts.first
                : null,
          },
          dragAnchorStrategy: pointerDragAnchorStrategy,
          feedback: Transform.translate(
            offset: const Offset(-12, -12),
            child: Container(
              width: 24,
              height: 24,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: const Color(0xFF6366F1),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFF6366F1).withValues(alpha: 0.5),
                    blurRadius: 12,
                    spreadRadius: 2,
                  ),
                ],
              ),
            ),
          ),
          child: AnimatedScale(
            scale: _isDragging ? 1.2 : 1.0,
            duration: const Duration(milliseconds: 150),
            child: const ConnectorWidget(isInput: false),
          ),
        ),
      ),
    );
  }
}

class InputOutputItem extends StatelessWidget {
  const InputOutputItem(this.entry, {this.isInput = true, super.key});

  final MapEntry entry;
  final bool isInput;

  @override
  Widget build(BuildContext context) {
    final color = isInput ? NodeHelper.inputColor : NodeHelper.outputColor;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 4,
          height: 4,
          decoration: BoxDecoration(shape: BoxShape.circle, color: color),
        ),
        const HorizontalSpacer(6),
        Flexible(
          child: Text(
            entry.key,
            style: style(
              size: 11,
              color: MyColors.textDarkGrey,
              weight: FontWeight.w500,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}

/// Widget for displaying a port label with data type indicator.
class _PortLabel extends StatelessWidget {
  const _PortLabel({required this.port, required this.isInput});

  final PortDefinition port;
  final bool isInput;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final portColor = port.dataType.color;

    return Row(
      mainAxisSize: MainAxisSize.min,
      mainAxisAlignment: isInput
          ? MainAxisAlignment.start
          : MainAxisAlignment.end,
      children: [
        // Data type indicator
        Container(
          width: 6,
          height: 6,
          decoration: BoxDecoration(shape: BoxShape.circle, color: portColor),
        ),
        const SizedBox(width: 4),
        // Port label
        Flexible(
          child: Text(
            port.label,
            style: style(
              size: 10,
              color: colors.textSecondary,
              weight: FontWeight.w500,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
        const SizedBox(width: 4),
        // Data type badge
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
          decoration: BoxDecoration(
            color: portColor.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(3),
          ),
          child: Text(
            port.dataType.displayName,
            style: style(size: 8, color: portColor, weight: FontWeight.w600),
          ),
        ),
      ],
    );
  }
}
