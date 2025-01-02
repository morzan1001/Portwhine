import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/canvas_cubit.dart';
import 'package:portwhine/blocs/single_pipeline/pipeline_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/global.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/canvas_model.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineControls extends StatelessWidget {
  const PipelineControls({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineCubit, PipelineModel>(
      builder: (context, state) {
        return Row(
          children: [
            // back button
            Container(
              height: 56,
              width: 56,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: MyColors.white,
                boxShadow: [
                  BoxShadow(
                    blurRadius: 6,
                    spreadRadius: 1,
                    color: MyColors.black.withOpacity(0.04),
                  ),
                ],
              ),
              child: InkWell(
                child: const Icon(
                  Icons.arrow_back,
                  color: MyColors.black,
                  size: 20,
                ),
                onTap: () => pop(context),
              ),
            ),
            const Spacer(),

            // buttons
            Container(
              height: 56,
              padding: const EdgeInsets.symmetric(horizontal: 32),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: MyColors.white,
                boxShadow: [
                  BoxShadow(
                    blurRadius: 6,
                    spreadRadius: 1,
                    color: MyColors.black.withOpacity(0.04),
                  ),
                ],
              ),
              child: Center(
                child: Text(
                  'View Results',
                  style: style(
                    color: MyColors.textDarkGrey,
                    weight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const HorizontalSpacer(16),
            Container(
              height: 56,
              padding: const EdgeInsets.symmetric(horizontal: 32),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: MyColors.white,
                boxShadow: [
                  BoxShadow(
                    blurRadius: 6,
                    spreadRadius: 1,
                    color: MyColors.black.withOpacity(0.04),
                  ),
                ],
              ),
              child: Center(
                child: Text(
                  'Save',
                  style: style(
                    color: MyColors.textDarkGrey,
                    weight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const HorizontalSpacer(16),

            // controls
            BlocBuilder<CanvasCubit, CanvasModel>(
              builder: (context, state) {
                return Container(
                  height: 56,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    color: MyColors.white,
                    boxShadow: [
                      BoxShadow(
                        blurRadius: 6,
                        spreadRadius: 1,
                        color: MyColors.black.withOpacity(0.04),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.pause_outlined,
                        color: MyColors.black,
                        size: 20,
                      ),
                      const HorizontalSpacer(20),
                      const Icon(
                        Icons.play_arrow_outlined,
                        color: MyColors.black,
                        size: 20,
                      ),
                      const HorizontalSpacer(20),
                      Container(
                        height: 32,
                        width: 1,
                        color: MyColors.darkGrey,
                      ),
                      const HorizontalSpacer(20),
                      InkWell(
                        child: const Icon(
                          Icons.info_outline,
                          color: MyColors.black,
                          size: 20,
                        ),
                        onTap: () {
                          Scaffold.of(context).openEndDrawer();
                        },
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
