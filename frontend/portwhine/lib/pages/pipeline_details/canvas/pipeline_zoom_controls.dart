import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineZoomControls extends StatelessWidget {
  const PipelineZoomControls({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CanvasCubit, CanvasModel>(
      builder: (context, state) {
        return Container(
          height: 48,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(24),
            color: MyColors.white,
            boxShadow: [
              BoxShadow(
                blurRadius: 16,
                spreadRadius: 2,
                color: MyColors.black.withValues(alpha: 0.08),
              ),
            ],
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              _ZoomButton(
                icon: Icons.remove_rounded,
                onTap: () => BlocProvider.of<CanvasCubit>(context).zoom(false),
                enabled: state.zoom > 0.2,
              ),
              const HorizontalSpacer(8),
              AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(
                  color: const Color(0xFF6366F1).withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  '${(state.zoom * 100).toStringAsFixed(0)}%',
                  style: style(
                    color: const Color(0xFF6366F1),
                    weight: FontWeight.w600,
                    size: 13,
                  ),
                ),
              ),
              const HorizontalSpacer(8),
              _ZoomButton(
                icon: Icons.add_rounded,
                onTap: () => BlocProvider.of<CanvasCubit>(context).zoom(true),
                enabled: state.zoom < 3.0,
              ),
            ],
          ),
        );
      },
    );
  }
}

class _ZoomButton extends StatefulWidget {
  const _ZoomButton({
    required this.icon,
    required this.onTap,
    this.enabled = true,
  });

  final IconData icon;
  final VoidCallback onTap;
  final bool enabled;

  @override
  State<_ZoomButton> createState() => _ZoomButtonState();
}

class _ZoomButtonState extends State<_ZoomButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: widget.enabled
          ? SystemMouseCursors.click
          : SystemMouseCursors.forbidden,
      child: GestureDetector(
        onTap: widget.enabled ? widget.onTap : null,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          width: 32,
          height: 32,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: _isHovered && widget.enabled
                ? const Color(0xFF6366F1).withValues(alpha: 0.15)
                : Colors.transparent,
          ),
          child: Center(
            child: Icon(
              widget.icon,
              color: widget.enabled
                  ? (_isHovered
                      ? const Color(0xFF6366F1)
                      : MyColors.textDarkGrey)
                  : MyColors.darkGrey,
              size: 20,
            ),
          ),
        ),
      ),
    );
  }
}
