import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/nodes_connection_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/single_pipeline/single_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/pages/pipeline_details/widgets/shadow_container.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/toast.dart';
import 'package:url_launcher/url_launcher.dart';

class PipelineControls extends StatelessWidget {
  const PipelineControls({super.key});

  Future<void> _launchKibana(BuildContext context, String pipelineId) async {
    final kibanaUrl = Uri.parse(
      'https://kibana.portwhine.local/app/discover#/?_a=(query:(language:kuery,query:\'pipeline_id:"$pipelineId"\'))',
    );

    try {
      if (!await launchUrl(kibanaUrl, mode: LaunchMode.externalApplication)) {
        if (context.mounted) {
          showToast(context, 'Could not launch Kibana');
        }
      }
    } catch (e) {
      if (context.mounted) {
        showToast(context, 'Error launching Kibana: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineCubit, PipelineModel>(
      builder: (context, state) {
        final isRunning = state.status == kStatusRunning;

        return Row(
          children: [
            // Back button with icon
            ShadowContainer(
              onTap: () => pop(context),
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  const Icon(
                    Icons.arrow_back_rounded,
                    color: Color(0xFF6366F1),
                    size: 20,
                  ),
                  const HorizontalSpacer(8),
                  Text(
                    'Back',
                    style: style(
                      color: MyColors.textDarkGrey,
                      weight: FontWeight.w500,
                      size: 13,
                    ),
                  ),
                ],
              ),
            ),

            const Spacer(),

            // View Results button
            ShadowContainer(
              onTap: () => _launchKibana(context, state.id),
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Row(
                children: [
                  Icon(
                    Icons.analytics_outlined,
                    color: const Color(0xFF6366F1),
                    size: 18,
                  ),
                  const HorizontalSpacer(8),
                  Text(
                    'Results',
                    style: style(
                      color: MyColors.textDarkGrey,
                      weight: FontWeight.w600,
                      size: 13,
                    ),
                  ),
                ],
              ),
            ),
            const HorizontalSpacer(12),

            // Save button
            ShadowContainer(
              onTap: () {
                final pipeline = context.read<PipelineCubit>().state;
                final nodes = context.read<NodesCubit>().state;

                final updatedPipeline = pipeline.copyWith(
                  nodes: nodes,
                );

                context
                    .read<SinglePipelineBloc>()
                    .add(UpdatePipeline(updatedPipeline));
              },
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Row(
                children: [
                  const Icon(
                    Icons.save_outlined,
                    color: Color(0xFF10B981),
                    size: 18,
                  ),
                  const HorizontalSpacer(8),
                  Text(
                    'Save',
                    style: style(
                      color: MyColors.textDarkGrey,
                      weight: FontWeight.w600,
                      size: 13,
                    ),
                  ),
                ],
              ),
            ),
            const HorizontalSpacer(12),

            // Play/Pause controls
            BlocBuilder<CanvasCubit, CanvasModel>(
              builder: (context, canvasState) {
                return Container(
                  height: 48,
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    color: MyColors.white,
                    boxShadow: [
                      BoxShadow(
                        blurRadius: 10,
                        spreadRadius: 1,
                        color: MyColors.black.withValues(alpha: 0.06),
                        offset: const Offset(0, 3),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      _ControlButton(
                        icon: Icons.stop_rounded,
                        color: const Color(0xFFEF4444),
                        isActive: isRunning,
                        onTap: isRunning
                            ? () {
                                context
                                    .read<StartStopPipelineBloc>()
                                    .add(StopPipeline(state.id));
                              }
                            : null,
                        tooltip: 'Stop',
                      ),
                      const HorizontalSpacer(4),
                      _ControlButton(
                        icon: Icons.play_arrow_rounded,
                        color: const Color(0xFF10B981),
                        isActive: !isRunning,
                        onTap: !isRunning
                            ? () {
                                context
                                    .read<StartStopPipelineBloc>()
                                    .add(StartPipeline(state.id));
                              }
                            : null,
                        tooltip: 'Start',
                      ),
                      const HorizontalSpacer(8),
                      Container(
                        height: 24,
                        width: 1,
                        color: MyColors.darkGrey.withValues(alpha: 0.5),
                      ),
                      const HorizontalSpacer(8),
                      _ControlButton(
                        icon: Icons.info_outline_rounded,
                        color: const Color(0xFF6366F1),
                        isActive: true,
                        onTap: () {
                          Scaffold.of(context).openEndDrawer();
                        },
                        tooltip: 'Pipeline Info',
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        );
      },
    );
  }
}

class _ControlButton extends StatefulWidget {
  const _ControlButton({
    required this.icon,
    required this.color,
    required this.isActive,
    required this.onTap,
    this.tooltip,
  });

  final IconData icon;
  final Color color;
  final bool isActive;
  final VoidCallback? onTap;
  final String? tooltip;

  @override
  State<_ControlButton> createState() => _ControlButtonState();
}

class _ControlButtonState extends State<_ControlButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final canTap = widget.isActive && widget.onTap != null;

    return Tooltip(
      message: widget.tooltip ?? '',
      child: MouseRegion(
        onEnter: (_) => setState(() => _isHovered = true),
        onExit: (_) => setState(() => _isHovered = false),
        cursor: canTap ? SystemMouseCursors.click : SystemMouseCursors.basic,
        child: GestureDetector(
          onTap: widget.onTap,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(8),
              color: _isHovered && canTap
                  ? widget.color.withValues(alpha: 0.15)
                  : Colors.transparent,
            ),
            child: Center(
              child: Icon(
                widget.icon,
                color: canTap
                    ? (_isHovered
                        ? widget.color
                        : widget.color.withValues(alpha: 0.7))
                    : MyColors.darkGrey,
                size: 22,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
