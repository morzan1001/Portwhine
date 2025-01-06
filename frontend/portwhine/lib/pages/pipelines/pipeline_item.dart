import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipelines_status/pipelines_status_bloc.dart';
import 'package:portwhine/blocs/pipelines/start_stop_pipeline/start_stop_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/constants.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/router/router.gr.dart';
import 'package:portwhine/widgets/icon_button.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/text.dart';

class PipelineItem extends StatelessWidget {
  const PipelineItem(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () {
        AutoRouter.of(context).navigate(
          PipelineDetailsRoute(id: model.id, model: model),
        );
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          color: MyColors.white,
        ),
        child: Row(
          children: [
            // name
            Expanded(
              flex: 1,
              child: Heading(model.name),
            ),
            const HorizontalSpacer(12),

            // status
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                const SmallText('STATUS', smaller: true),
                const VerticalSpacer(2),
                StatusText(model),
              ],
            ),
            const HorizontalSpacer(24),

            // start / stop button
            StartStopButton(model),
            const HorizontalSpacer(8),

            // delete button
            DeleteButton(model),
          ],
        ),
      ),
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
                  size: 16,
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
