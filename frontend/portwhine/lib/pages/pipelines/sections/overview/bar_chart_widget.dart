import 'dart:math';

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/bloc/workflows/workflows_number/workflows_number_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/loading_indicator.dart';
import 'package:intl/intl.dart';
import 'package:shimmer/shimmer.dart';

class BarChartWidget extends StatefulWidget {
  const BarChartWidget({super.key});

  @override
  State<BarChartWidget> createState() => _BarChartWidgetState();
}

class _BarChartWidgetState extends State<BarChartWidget> {
  @override
  Widget build(BuildContext context) {
    return BlocBuilder<WorkflowsNumberBloc, WorkflowsNumberState>(
      builder: (context, state) {
        if (state is WorkflowsNumberLoading) {
          return Shimmer.fromColors(
            baseColor: CustomColors.greyVar,
            highlightColor: CustomColors.grey,
            child: Row(
              children: [
                ...List.generate(
                  8,
                  (i) => Expanded(
                    child: Column(
                      children: [
                        Container(
                          margin: const EdgeInsets.symmetric(horizontal: 4),
                          height: 100,
                          width: 10,
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          );
        }

        if (state is WorkflowsNumberFailed) {
          return Text(state.error);
        }

        if (state is WorkflowsNumberLoaded) {
          final numbers = state.numbers.values;
          final maxNumber = numbers.reduce(max);
          final entries = state.numbers.entries.toList();

          return SizedBox(
            height: 110,
            child: BarChart(
              BarChartData(
                barTouchData: BarTouchData(
                  touchTooltipData: BarTouchTooltipData(
                    // tooltipBgColor: CustomColors.white,
                    tooltipHorizontalAlignment: FLHorizontalAlignment.right,
                    tooltipMargin: -20,
                  ),
                ),
                titlesData: FlTitlesData(
                  show: true,
                  rightTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  topTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  leftTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: maxNumber.toString().length * 10,
                      getTitlesWidget: (value, meta) {
                        return SideTitleWidget(
                          axisSide: meta.axisSide,
                          space: 4,
                          child: Text(
                            value.toInt().toString(),
                            style: style(size: 11),
                          ),
                        );
                      },
                      interval: maxNumber / 4,
                    ),
                  ),
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      getTitlesWidget: (value, meta) {
                        return SideTitleWidget(
                          axisSide: meta.axisSide,
                          space: 8,
                          child: Text(
                            DateFormat('MMM')
                                .format(entries[value ~/ 1].key)
                                .substring(0, 2),
                            style: style(size: 10),
                          ),
                        );
                      },
                      interval: 10,
                    ),
                  ),
                ),
                barGroups: List.generate(
                  numbers.length,
                  (i) => BarChartGroupData(
                    x: i,
                    barRods: [
                      BarChartRodData(
                        toY: entries[i].value.toDouble(),
                        color: CustomColors.sec,
                        width: 10,
                        backDrawRodData: BackgroundBarChartRodData(
                          show: true,
                          toY: maxNumber.toDouble(),
                          color: CustomColors.sec.withOpacity(0),
                        ),
                      ),
                    ],
                  ),
                ),
                borderData: FlBorderData(show: false),
                gridData: const FlGridData(show: false),
              ),
            ),
          );
        }

        return const LoadingIndicator();
      },
    );
  }
}
