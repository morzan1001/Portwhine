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
              InkWell(
                onTap: () {
                  BlocProvider.of<CanvasCubit>(context).zoom(false);
                },
                child: const Icon(
                  Icons.remove,
                  color: MyColors.black,
                  size: 20,
                ),
              ),
              const HorizontalSpacer(20),
              Text(
                '${(state.zoom * 100).toStringAsFixed(0)}%',
                style: style(
                  color: MyColors.textDarkGrey,
                  weight: FontWeight.w600,
                ),
              ),
              const HorizontalSpacer(20),
              InkWell(
                onTap: () {
                  BlocProvider.of<CanvasCubit>(context).zoom(true);
                },
                child: const Icon(
                  Icons.add,
                  color: MyColors.black,
                  size: 20,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
