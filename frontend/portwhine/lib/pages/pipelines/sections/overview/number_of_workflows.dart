import 'package:flutter/material.dart';
import 'package:frontend/global/colors.dart';
import 'package:frontend/global/text_style.dart';
import 'package:frontend/pages/pipelines/sections/overview/bar_chart_widget.dart';
import 'package:frontend/widgets/spacer.dart';

class NumberOfPipelines extends StatelessWidget {
  const NumberOfPipelines({super.key});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: 24,
          vertical: 16,
        ),
        decoration: BoxDecoration(
          color: CustomColors.white,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Number of Pipelines',
              style: style(
                size: 14,
                color: CustomColors.textDark,
                weight: FontWeight.w600,
              ),
            ),
            const VerticalSpacer(24),
            const Spacer(),
            const BarChartWidget(),
          ],
        ),
      ),
    );
  }
}
