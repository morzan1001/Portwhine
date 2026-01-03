import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/helpers.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineInfoDrawer extends StatelessWidget {
  const PipelineInfoDrawer({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineCubit, PipelineModel>(
      builder: (context, pipeline) {
        final statusColor = StatusHelper.getStatusColor(pipeline.status);

        return Drawer(
          width: 360,
          backgroundColor: MyColors.white,
          shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.horizontal(left: Radius.circular(20)),
          ),
          child: SafeArea(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        const Color(0xFF6366F1).withValues(alpha: 0.1),
                        const Color(0xFF6366F1).withValues(alpha: 0.05),
                      ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: const Color(0xFF6366F1)
                                  .withValues(alpha: 0.15),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: const Icon(
                              Icons.account_tree_rounded,
                              color: Color(0xFF6366F1),
                              size: 24,
                            ),
                          ),
                          const Spacer(),
                          IconButton(
                            onPressed: () => Navigator.of(context).pop(),
                            icon: const Icon(Icons.close_rounded),
                            color: MyColors.textDarkGrey,
                          ),
                        ],
                      ),
                      const VerticalSpacer(16),
                      Text(
                        pipeline.name.isNotEmpty ? pipeline.name : 'Pipeline',
                        style: style(
                          color: MyColors.black,
                          size: 22,
                          weight: FontWeight.w700,
                        ),
                      ),
                      const VerticalSpacer(8),
                      Row(
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: statusColor,
                              boxShadow: [
                                BoxShadow(
                                  color: statusColor.withValues(alpha: 0.5),
                                  blurRadius: 6,
                                  spreadRadius: 1,
                                ),
                              ],
                            ),
                          ),
                          const HorizontalSpacer(8),
                          Text(
                            pipeline.status.isNotEmpty
                                ? pipeline.status
                                : 'Unknown',
                            style: style(
                              color: statusColor,
                              size: 14,
                              weight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Content
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildInfoSection(
                          'Pipeline Details',
                          Icons.info_outline_rounded,
                          [
                            _InfoItem('ID', pipeline.id),
                            _InfoItem('Name', pipeline.name),
                            _InfoItem('Status', pipeline.status),
                            _InfoItem('Nodes', '${pipeline.nodes.length}'),
                            _InfoItem(
                                'Connections', '${pipeline.edges.length}'),
                          ],
                        ),
                        const VerticalSpacer(24),
                        _buildInfoSection(
                          'Nodes',
                          Icons.extension_rounded,
                          pipeline.nodes
                              .map((node) => _InfoItem(
                                  node.name,
                                  node.id.substring(0,
                                      node.id.length > 8 ? 8 : node.id.length)))
                              .toList(),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildInfoSection(String title, IconData icon, List<_InfoItem> items) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, color: const Color(0xFF6366F1), size: 18),
            const HorizontalSpacer(8),
            Text(
              title,
              style: style(
                color: MyColors.black,
                size: 16,
                weight: FontWeight.w600,
              ),
            ),
          ],
        ),
        const VerticalSpacer(12),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: MyColors.lightGrey,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            children: items.asMap().entries.map((entry) {
              final isLast = entry.key == items.length - 1;
              return Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        entry.value.label,
                        style: style(
                          color: MyColors.textDarkGrey,
                          size: 13,
                        ),
                      ),
                      Flexible(
                        child: Text(
                          entry.value.value,
                          style: style(
                            color: MyColors.black,
                            size: 13,
                            weight: FontWeight.w500,
                          ),
                          textAlign: TextAlign.end,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                  if (!isLast) ...[
                    const VerticalSpacer(12),
                    Container(
                      height: 1,
                      color: MyColors.darkGrey.withValues(alpha: 0.2),
                    ),
                    const VerticalSpacer(12),
                  ],
                ],
              );
            }).toList(),
          ),
        ),
      ],
    );
  }
}

class _InfoItem {
  final String label;
  final String value;

  _InfoItem(this.label, this.value);
}
