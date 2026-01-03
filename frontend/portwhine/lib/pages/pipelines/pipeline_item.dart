import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_status/pipelines_status_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/helpers.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/router/router.gr.dart';
import 'package:portwhine/widgets/icon_button.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineItem extends StatefulWidget {
  const PipelineItem(this.model, {super.key});

  final PipelineModel model;

  @override
  State<PipelineItem> createState() => _PipelineItemState();
}

class _PipelineItemState extends State<PipelineItem> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOutCubic,
        transform: Matrix4.diagonal3Values(
            _isHovered ? 1.01 : 1.0, _isHovered ? 1.01 : 1.0, 1.0),
        transformAlignment: Alignment.center,
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            borderRadius: BorderRadius.circular(16),
            onTap: () {
              AutoRouter.of(context)
                  .navigate(PipelineDetailsRoute(id: widget.model.id));
            },
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: double.infinity,
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(16),
                color: MyColors.white,
                border: Border.all(
                  color: _isHovered
                      ? const Color(0xFF6366F1).withValues(alpha: 0.5)
                      : Colors.transparent,
                  width: 2,
                ),
                boxShadow: [
                  BoxShadow(
                    color: _isHovered
                        ? const Color(0xFF6366F1).withValues(alpha: 0.15)
                        : MyColors.black.withValues(alpha: 0.05),
                    blurRadius: _isHovered ? 20 : 8,
                    spreadRadius: _isHovered ? 1 : 0,
                    offset: Offset(0, _isHovered ? 6 : 2),
                  ),
                ],
              ),
              child: Row(
                children: [
                  // Status indicator dot
                  StatusIndicator(widget.model),
                  const HorizontalSpacer(16),

                  // name
                  Expanded(
                    flex: 1,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          widget.model.name,
                          style: style(
                            size: 16,
                            weight: FontWeight.w600,
                            color: MyColors.black,
                          ),
                        ),
                        const VerticalSpacer(4),
                        Text(
                          widget.model.id,
                          style: style(
                            size: 12,
                            color: MyColors.textDarkGrey,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const HorizontalSpacer(16),

                  // status text
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        'STATUS',
                        style: style(
                          size: 10,
                          weight: FontWeight.w600,
                          color: MyColors.textDarkGrey,
                          spacing: 1.2,
                        ),
                      ),
                      const VerticalSpacer(4),
                      StatusText(widget.model),
                    ],
                  ),
                  const HorizontalSpacer(24),

                  // start / stop button
                  StartStopButton(widget.model),
                  const HorizontalSpacer(8),

                  // delete button
                  DeleteButton(widget.model),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class StatusIndicator extends StatelessWidget {
  const StatusIndicator(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinesStatusBloc, PipelinesStatusState>(
      builder: (context, statusState) {
        Color color = MyColors.grey;
        bool isAnimated = false;

        if (statusState is PipelinesStatusUpdated) {
          final pipeline = statusState.pipelines.singleWhere(
            (e) => e.id == model.id,
          );
          final status = pipeline.status;
          color = StatusHelper.getStatusColor(status);
          isAnimated = StatusHelper.isRunning(status);
        }

        return Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: color,
            boxShadow: isAnimated
                ? [
                    BoxShadow(
                      color: color.withValues(alpha: 0.5),
                      blurRadius: 8,
                      spreadRadius: 2,
                    ),
                  ]
                : [],
          ),
        );
      },
    );
  }
}

class StatusText extends StatelessWidget {
  const StatusText(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinesStatusBloc, PipelinesStatusState>(
      builder: (context, statusState) {
        final updated = statusState is PipelinesStatusUpdated;

        if (updated) {
          final pipeline = statusState.pipelines.singleWhere(
            (e) => e.id == model.id,
          );

          final status = pipeline.status;

          return Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              if (status == kStatusError)
                const Icon(
                  Icons.report_outlined,
                  color: MyColors.red,
                  size: 20,
                ),
              if (status == kStatusError) const HorizontalSpacer(4),
              Text(
                status,
                style: style(
                  size: 15,
                  color: status == kStatusRunning
                      ? MyColors.green
                      : status == kStatusError
                          ? MyColors.red
                          : MyColors.black,
                ),
              ),
            ],
          );
        }

        return const SizedBox.shrink();
      },
    );
  }
}

class StartStopButton extends StatelessWidget {
  const StartStopButton(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinesStatusBloc, PipelinesStatusState>(
      builder: (context, statusState) {
        final updated = statusState is PipelinesStatusUpdated;

        if (updated) {
          final pipeline = statusState.pipelines.singleWhere(
            (e) => e.id == model.id,
          );

          final status = pipeline.status;
          final start = status == kStatusUnknown || status == kStatusStopped;

          return BlocBuilder<StartStopPipelineBloc, StartStopPipelineState>(
            builder: (context, startStopState) {
              return MyIconButton(
                start ? Icons.play_arrow : Icons.pause,
                showLoading: startStopState is StartStopPipelineStarted &&
                    startStopState.id == model.id,
                onTap: () {
                  BlocProvider.of<StartStopPipelineBloc>(context).add(
                    start ? StartPipeline(model.id) : StopPipeline(model.id),
                  );
                },
              );
            },
          );
        }

        return const SizedBox.shrink();
      },
    );
  }
}

class DeleteButton extends StatelessWidget {
  const DeleteButton(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<DeletePipelineBloc, DeletePipelineState>(
      builder: (context, deleteState) {
        return MyIconButton(
          Icons.delete,
          iconColor: MyColors.red,
          showLoading: deleteState is DeletePipelineStarted &&
              deleteState.id == model.id,
          onTap: () {
            BlocProvider.of<DeletePipelineBloc>(context).add(
              DeletePipeline(model.id),
            );
          },
        );
      },
    );
  }
}
