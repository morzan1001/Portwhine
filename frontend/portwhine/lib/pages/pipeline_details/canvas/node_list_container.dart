import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeListContainer extends StatelessWidget {
  const NodeListContainer({
    required this.title,
    required this.child,
    this.icon,
    super.key,
  });

  final String title;
  final Widget child;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    final isWorkers = title.toLowerCase() == 'workers';
    final headerColor =
        isWorkers ? const Color(0xFF6366F1) : const Color(0xFF10B981);

    return Container(
      width: 300,
      decoration: BoxDecoration(
        color: MyColors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: MyColors.black.withValues(alpha: 0.08),
            blurRadius: 20,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  headerColor.withValues(alpha: 0.1),
                  headerColor.withValues(alpha: 0.05),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(16),
              ),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: headerColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    isWorkers ? Icons.extension_rounded : Icons.bolt_rounded,
                    color: headerColor,
                    size: 18,
                  ),
                ),
                const HorizontalSpacer(12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: style(
                          color: MyColors.black,
                          size: 15,
                          weight: FontWeight.w600,
                        ),
                      ),
                      const VerticalSpacer(2),
                      Text(
                        isWorkers ? 'Drag to canvas' : 'Start pipeline',
                        style: style(
                          color: MyColors.textDarkGrey,
                          size: 11,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // Divider
          Container(
            height: 1,
            color: headerColor.withValues(alpha: 0.1),
          ),
          // Content
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: child,
            ),
          ),
        ],
      ),
    );
  }
}
