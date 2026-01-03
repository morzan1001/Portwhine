import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/helpers.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeDetails extends StatefulWidget {
  const NodeDetails(this.model, {super.key});

  final NodeModel model;

  @override
  State<NodeDetails> createState() => _NodeDetailsState();
}

class _NodeDetailsState extends State<NodeDetails>
    with SingleTickerProviderStateMixin {
  late Map<String, dynamic> _config;
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();
    _config = Map.from(widget.model.config);

    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );

    _fadeAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeOut),
    );

    _slideAnimation =
        Tween<Offset>(begin: const Offset(0, 0.1), end: Offset.zero).animate(
          CurvedAnimation(
            parent: _animationController,
            curve: Curves.easeOutCubic,
          ),
        );

    _animationController.forward();
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant NodeDetails oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.model != widget.model) {
      _config = Map.from(widget.model.config);
    }
  }

  void _updateConfig(String key, dynamic value) {
    setState(() {
      _config[key] = value;
    });

    final updatedNode = widget.model.copyWith(config: _config);
    context.read<NodesCubit>().updateNode(updatedNode);
    context.read<SelectedNodeCubit>().setNode(updatedNode);
  }

  Color _getNodeTypeColor() => NodeHelper.getNodeColor(widget.model.name);

  IconData _getNodeIcon() => NodeHelper.getNodeIcon(widget.model.name);

  Widget _buildConfigField(String key, dynamic value) {
    final nodeColor = _getNodeTypeColor();

    if (value is bool) {
      return Container(
        margin: const EdgeInsets.only(bottom: 12),
        decoration: BoxDecoration(
          color: MyColors.lightGrey,
          borderRadius: BorderRadius.circular(12),
        ),
        child: SwitchListTile(
          title: Text(
            key,
            style: style(color: MyColors.black, weight: FontWeight.w500),
          ),
          value: value,
          onChanged: (newValue) => _updateConfig(key, newValue),
          activeTrackColor: nodeColor.withValues(alpha: 0.5),
          thumbColor: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.selected)) {
              return nodeColor;
            }
            return null;
          }),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      );
    } else if (value is int) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: TextFormField(
          initialValue: value.toString(),
          decoration: InputDecoration(
            labelText: key,
            labelStyle: style(color: MyColors.textDarkGrey),
            filled: true,
            fillColor: MyColors.lightGrey,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide.none,
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: nodeColor, width: 2),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
          ),
          keyboardType: TextInputType.number,
          onChanged: (newValue) {
            final intValue = int.tryParse(newValue);
            if (intValue != null) {
              _updateConfig(key, intValue);
            }
          },
        ),
      );
    } else if (value is String) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: TextFormField(
          initialValue: value,
          decoration: InputDecoration(
            labelText: key,
            labelStyle: style(color: MyColors.textDarkGrey),
            filled: true,
            fillColor: MyColors.lightGrey,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide.none,
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: nodeColor, width: 2),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
          ),
          onChanged: (newValue) => _updateConfig(key, newValue),
        ),
      );
    } else if (value is List) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: TextFormField(
          initialValue: value.join(', '),
          decoration: InputDecoration(
            labelText: key,
            labelStyle: style(color: MyColors.textDarkGrey),
            filled: true,
            fillColor: MyColors.lightGrey,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide.none,
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: nodeColor, width: 2),
            ),
            helperText: 'Comma-separated values',
            helperStyle: style(color: MyColors.textLightGrey, size: 12),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
          ),
          onChanged: (newValue) {
            final list = newValue.split(',').map((e) => e.trim()).toList();
            if (value.isNotEmpty && value.first is int) {
              final intList = list
                  .map((e) => int.tryParse(e))
                  .whereType<int>()
                  .toList();
              _updateConfig(key, intList);
            } else {
              _updateConfig(key, list);
            }
          },
        ),
      );
    }
    return Text('Unsupported type for $key: ${value.runtimeType}');
  }

  @override
  Widget build(BuildContext context) {
    final nodeColor = _getNodeTypeColor();

    return FadeTransition(
      opacity: _fadeAnimation,
      child: SlideTransition(
        position: _slideAnimation,
        child: Container(
          width: 480,
          constraints: const BoxConstraints(maxHeight: 600),
          decoration: BoxDecoration(
            color: MyColors.white,
            borderRadius: BorderRadius.circular(20),
            boxShadow: [
              BoxShadow(
                color: nodeColor.withValues(alpha: 0.1),
                blurRadius: 30,
                spreadRadius: 5,
              ),
              BoxShadow(
                color: MyColors.black.withValues(alpha: 0.1),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      nodeColor.withValues(alpha: 0.1),
                      nodeColor.withValues(alpha: 0.05),
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: const BorderRadius.vertical(
                    top: Radius.circular(20),
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: nodeColor.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Icon(_getNodeIcon(), color: nodeColor, size: 24),
                    ),
                    const HorizontalSpacer(16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            widget.model.definition?.name ?? widget.model.name,
                            style: style(
                              color: MyColors.black,
                              size: 18,
                              weight: FontWeight.w600,
                            ),
                          ),
                          const VerticalSpacer(4),
                          Text(
                            'Node Configuration',
                            style: style(
                              color: MyColors.textDarkGrey,
                              size: 13,
                            ),
                          ),
                        ],
                      ),
                    ),
                    _buildActionButton(
                      icon: Icons.delete_outline_rounded,
                      color: const Color(0xFFEF4444),
                      onTap: () {
                        context.read<NodesCubit>().removeNode(widget.model.id);
                        context.read<SelectedNodeCubit>().removeNode();
                        Navigator.of(context).pop();
                      },
                    ),
                    const HorizontalSpacer(8),
                    _buildActionButton(
                      icon: Icons.close_rounded,
                      color: MyColors.textDarkGrey,
                      onTap: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              // Content
              Flexible(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (_config.isNotEmpty) ...[
                        _buildSectionTitle(
                          'Konfiguration',
                          Icons.settings_rounded,
                        ),
                        const VerticalSpacer(12),
                        ..._config.entries.map(
                          (e) => _buildConfigField(e.key, e.value),
                        ),
                        const VerticalSpacer(20),
                      ],
                      _buildSectionTitle(
                        'Eingänge & Ausgänge',
                        Icons.swap_horiz_rounded,
                      ),
                      const VerticalSpacer(12),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(
                            child: _buildIOSection(
                              'Inputs',
                              widget.model.inputs.keys.toList(),
                              const Color(0xFF10B981),
                              Icons.arrow_forward_rounded,
                            ),
                          ),
                          const HorizontalSpacer(16),
                          Expanded(
                            child: _buildIOSection(
                              'Outputs',
                              widget.model.outputs.keys.toList(),
                              const Color(0xFF6366F1),
                              Icons.arrow_back_rounded,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildActionButton({
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: color, size: 20),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, color: MyColors.textDarkGrey, size: 18),
        const HorizontalSpacer(8),
        Text(
          title,
          style: style(
            color: MyColors.black,
            size: 15,
            weight: FontWeight.w600,
          ),
        ),
      ],
    );
  }

  Widget _buildIOSection(
    String title,
    List<String> items,
    Color color,
    IconData icon,
  ) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: color, size: 16),
              const HorizontalSpacer(6),
              Text(
                title,
                style: style(color: color, size: 13, weight: FontWeight.w600),
              ),
            ],
          ),
          const VerticalSpacer(10),
          if (items.isEmpty)
            Text('Keine', style: style(color: MyColors.textLightGrey, size: 12))
          else
            ...items.map(
              (item) => Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: Row(
                  children: [
                    Container(
                      width: 6,
                      height: 6,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: color,
                      ),
                    ),
                    const HorizontalSpacer(8),
                    Expanded(
                      child: Text(
                        item,
                        style: style(
                          color: MyColors.black,
                          size: 13,
                          weight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }
}

Future showNodeDetailsDialog(BuildContext context, NodeModel model) async {
  return await showGeneralDialog(
    context: context,
    barrierColor: Colors.black54,
    barrierDismissible: true,
    barrierLabel: 'Node Details',
    transitionDuration: const Duration(milliseconds: 300),
    pageBuilder: (context, animation, secondaryAnimation) {
      return Center(
        child: Material(color: Colors.transparent, child: NodeDetails(model)),
      );
    },
    transitionBuilder: (context, animation, secondaryAnimation, child) {
      return FadeTransition(
        opacity: animation,
        child: ScaleTransition(
          scale: Tween<double>(begin: 0.95, end: 1.0).animate(
            CurvedAnimation(parent: animation, curve: Curves.easeOutCubic),
          ),
          child: child,
        ),
      );
    },
  );
}
