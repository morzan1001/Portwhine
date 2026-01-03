import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/bloc_listeners.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/global/theme.dart';
import 'package:portwhine/pages/pipelines/change_page_section.dart';
import 'package:portwhine/pages/pipelines/pipeline_item.dart';
import 'package:portwhine/pages/write_pipeline/write_pipeline.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/svg_icon.dart';
import 'package:portwhine/widgets/theme_toggle.dart';
import 'package:shimmer/shimmer.dart';

@RoutePage()
class PipelinesPage extends StatelessWidget {
  const PipelinesPage({super.key});

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;

    return MultiBlocListener(
      listeners: BlocListeners.pipelinesListener,
      child: SafeArea(
        child: Scaffold(
          backgroundColor: colors.surfaceVariant,
          body: Column(
            children: [
              // Modern gradient header
              Container(
                width: double.infinity,
                padding: const EdgeInsets.fromLTRB(32, 32, 32, 24),
                decoration: BoxDecoration(
                  gradient: AppColors.primaryGradient,
                  boxShadow: [
                    BoxShadow(
                      color: colors.primary.withValues(alpha: 0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 8),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Logo and theme toggle
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const SvgIcon(
                          icon: 'logo',
                          size: 56,
                          color: Colors.white,
                        ),
                        const ThemeToggleButton(size: 44),
                      ],
                    ),
                    const VerticalSpacer(24),

                    // Heading and add button
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Pipelines',
                              style: style(
                                size: 28,
                                weight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                            const VerticalSpacer(4),
                            Text(
                              'Manage your security scanning pipelines',
                              style: style(
                                size: 14,
                                color: Colors.white.withValues(alpha: 0.8),
                              ),
                            ),
                          ],
                        ),
                        _AddPipelineButton(
                          onTap: () => showWritePipelineDialog(context),
                        ),
                      ],
                    ),
                  ],
                ),
              ),

              // Pipeline list
              Expanded(
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 24,
                  ),
                  child: Column(
                    children: [
                      Expanded(
                        child: BlocBuilder<PipelinesListBloc, PipelinesListState>(
                          builder: (context, state) {
                            if (state is PipelinesListLoading) {
                              return Shimmer.fromColors(
                                baseColor: MyColors.grey,
                                highlightColor: MyColors.lightGrey,
                                child: Column(
                                  children: List.generate(
                                    4,
                                    (i) => Container(
                                      margin: const EdgeInsets.only(bottom: 16),
                                      height: 88,
                                      width: double.infinity,
                                      decoration: BoxDecoration(
                                        borderRadius: BorderRadius.circular(16),
                                        color: MyColors.white,
                                      ),
                                    ),
                                  ),
                                ),
                              );
                            }

                            if (state is PipelinesListFailed) {
                              return Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(
                                      Icons.error_outline,
                                      size: 48,
                                      color: MyColors.red.withValues(
                                        alpha: 0.7,
                                      ),
                                    ),
                                    const VerticalSpacer(16),
                                    Text(
                                      state.error,
                                      style: style(
                                        color: MyColors.textDarkGrey,
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            }

                            if (state is PipelinesListLoaded) {
                              final pipelines = state.pipelines;

                              if (pipelines.isEmpty) {
                                return Center(
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.inbox_outlined,
                                        size: 64,
                                        color: MyColors.grey.withValues(
                                          alpha: 0.5,
                                        ),
                                      ),
                                      const VerticalSpacer(16),
                                      Text(
                                        'No pipelines yet',
                                        style: style(
                                          size: 18,
                                          weight: FontWeight.w600,
                                          color: MyColors.textDarkGrey,
                                        ),
                                      ),
                                      const VerticalSpacer(8),
                                      Text(
                                        'Create your first pipeline to get started',
                                        style: style(
                                          size: 14,
                                          color: MyColors.grey,
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              }

                              return ListView.separated(
                                separatorBuilder: (context, index) =>
                                    const VerticalSpacer(16),
                                itemCount: pipelines.length,
                                itemBuilder: (context, i) =>
                                    PipelineItem(pipelines[i]),
                              );
                            }

                            return const SizedBox.shrink();
                          },
                        ),
                      ),
                      const VerticalSpacer(16),

                      // Pagination
                      const PaginationSection(),
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
}

class _AddPipelineButton extends StatefulWidget {
  const _AddPipelineButton({required this.onTap});

  final VoidCallback onTap;

  @override
  State<_AddPipelineButton> createState() => _AddPipelineButtonState();
}

class _AddPipelineButtonState extends State<_AddPipelineButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        transform: Matrix4.diagonal3Values(
          _isHovered ? 1.05 : 1.0,
          _isHovered ? 1.05 : 1.0,
          1.0,
        ),
        transformAlignment: Alignment.center,
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: widget.onTap,
            borderRadius: BorderRadius.circular(12),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: Colors.white,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(
                      alpha: _isHovered ? 0.2 : 0.1,
                    ),
                    blurRadius: _isHovered ? 16 : 8,
                    offset: Offset(0, _isHovered ? 6 : 4),
                  ),
                ],
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 24,
                    height: 24,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: const LinearGradient(
                        colors: [Color(0xFF6366F1), Color(0xFF8B5CF6)],
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF6366F1).withValues(alpha: 0.4),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: const Icon(Icons.add, size: 16, color: Colors.white),
                  ),
                  const HorizontalSpacer(12),
                  Text(
                    'New Pipeline',
                    style: style(
                      size: 14,
                      weight: FontWeight.w600,
                      color: const Color(0xFF6366F1),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
