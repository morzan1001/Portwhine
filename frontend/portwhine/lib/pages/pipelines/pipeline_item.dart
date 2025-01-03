import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/delete_pipeline/delete_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
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
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'STATUS',
                  style: style(
                    color: MyColors.textLightGrey,
                    weight: FontWeight.w400,
                    size: 12,
                  ),
                ),
                Row(
                  children: [
                    Container(height: 8, width: 18, color: MyColors.green),
                    const HorizontalSpacer(2),
                    Container(height: 8, width: 18, color: MyColors.green),
                    const HorizontalSpacer(2),
                    Container(height: 8, width: 18, color: MyColors.green),
                  ],
                ),
                Text(
                  'RUNNING',
                  style: style(
                    color: MyColors.textDarkGrey,
                    weight: FontWeight.w600,
                    size: 12,
                  ),
                ),
              ],
            ),
            const HorizontalSpacer(24),

            // play / stop button
            MyIconButton(
              Icons.play_arrow,
              onTap: () {},
            ),
            const HorizontalSpacer(8),

            // delete button
            BlocBuilder<DeletePipelineBloc, DeletePipelineState>(
              builder: (context, state) {
                return MyIconButton(
                  Icons.delete,
                  iconColor: MyColors.red,
                  showLoading:
                      state is DeletePipelineStarted && state.id == model.id,
                  onTap: () {
                    BlocProvider.of<DeletePipelineBloc>(context).add(
                      DeletePipeline(model.id),
                    );
                  },
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
