import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/node_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeDetails extends StatefulWidget {
  const NodeDetails(this.model, {super.key});

  final NodeModel model;

  @override
  State<NodeDetails> createState() => _NodeDetailsState();
}

class _NodeDetailsState extends State<NodeDetails> {
  late Map<String, dynamic> _config;

  @override
  void initState() {
    super.initState();
    _config = Map.from(widget.model.config);
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

  Widget _buildConfigField(String key, dynamic value) {
    if (value is bool) {
      return SwitchListTile(
        title: Text(key, style: style(color: MyColors.black)),
        value: value,
        onChanged: (newValue) => _updateConfig(key, newValue),
        activeColor: MyColors.red,
      );
    } else if (value is int) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8.0),
        child: TextFormField(
          initialValue: value.toString(),
          decoration: InputDecoration(
            labelText: key,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
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
        padding: const EdgeInsets.symmetric(vertical: 8.0),
        child: TextFormField(
          initialValue: value,
          decoration: InputDecoration(
            labelText: key,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
          onChanged: (newValue) => _updateConfig(key, newValue),
        ),
      );
    } else if (value is List) {
      // Handle lists as comma-separated strings for now
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 8.0),
        child: TextFormField(
          initialValue: value.join(', '),
          decoration: InputDecoration(
            labelText: key,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            helperText: 'Comma separated values',
          ),
          onChanged: (newValue) {
            final list = newValue.split(',').map((e) => e.trim()).toList();
            if (value.isNotEmpty && value.first is int) {
              final intList =
                  list.map((e) => int.tryParse(e)).whereType<int>().toList();
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
    return Container(
      width: 500,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: MyColors.white,
        borderRadius: const BorderRadius.vertical(
          top: Radius.circular(16),
        ),
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  widget.model.name,
                  style: style(
                    color: MyColors.black,
                    size: 20,
                    weight: FontWeight.w600,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.delete, color: MyColors.red),
                  onPressed: () {
                    context.read<NodesCubit>().removeNode(widget.model.id);
                    context.read<SelectedNodeCubit>().removeNode();
                    Navigator.of(context).pop();
                  },
                ),
              ],
            ),
            const VerticalSpacer(16),
            if (_config.isNotEmpty) ...[
              Text(
                'Configuration',
                style: style(
                  color: MyColors.black,
                  size: 18,
                  weight: FontWeight.w500,
                ),
              ),
              const VerticalSpacer(8),
              ..._config.entries.map((e) => _buildConfigField(e.key, e.value)),
              const VerticalSpacer(16),
            ],
            Text(
              'I/O',
              style: style(
                color: MyColors.black,
                size: 18,
                weight: FontWeight.w500,
              ),
            ),
            const VerticalSpacer(16),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Inputs',
                        style: style(
                          color: MyColors.black,
                          size: 16,
                          weight: FontWeight.w500,
                        ),
                      ),
                      if (widget.model.inputs.isEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 8.0),
                          child:
                              Text("None", style: style(color: MyColors.grey)),
                        ),
                      ...List.generate(
                        widget.model.inputs.length,
                        (index) {
                          return Container(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            margin: const EdgeInsets.only(top: 12),
                            decoration: BoxDecoration(
                              color: MyColors.black.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Center(
                              child: Text(
                                widget.model.inputs.entries.toList()[index].key,
                                style: style(color: MyColors.black),
                              ),
                            ),
                          );
                        },
                      )
                    ],
                  ),
                ),
                const HorizontalSpacer(16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Outputs',
                        style: style(
                          color: MyColors.black,
                          size: 16,
                          weight: FontWeight.w500,
                        ),
                      ),
                      if (widget.model.outputs.isEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 8.0),
                          child:
                              Text("None", style: style(color: MyColors.grey)),
                        ),
                      ...List.generate(
                        widget.model.outputs.length,
                        (index) {
                          return Container(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            margin: const EdgeInsets.only(top: 12),
                            decoration: BoxDecoration(
                              color: MyColors.black.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Center(
                              child: Text(
                                widget.model.outputs.entries
                                    .toList()[index]
                                    .key,
                                style: style(color: MyColors.black),
                              ),
                            ),
                          );
                        },
                      )
                    ],
                  ),
                ),
              ],
            )
          ],
        ),
      ),
    );
  }
}

Future showNodeDetailsDialog(BuildContext context, NodeModel model) async {
  return await showDialog(
    context: context,
    barrierColor: Colors.transparent,
    builder: (context) {
      return Dialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        child: NodeDetails(model),
      );
    },
  );
}
