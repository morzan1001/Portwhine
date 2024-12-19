import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/pages/pipelines/sections/list/list.dart';
import 'package:portwhine/pages/pipelines/sections/overview/overview.dart';
import 'package:portwhine/widgets/spacer.dart';

@RoutePage()
class PipelinesPage extends StatelessWidget {
  const PipelinesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: CustomColors.greyLighter,
      child: Column(
        children: [
          Expanded(
            child: ListView(
              children: [
                Container(
                  color: CustomColors.greyLighter,
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Flexible(
                        child: Container(
                          constraints: const BoxConstraints(
                            maxWidth: 1200,
                          ),
                          padding: const EdgeInsets.all(16),
                          child: const Column(
                            children: [
                              PipelinesOverview(),
                              VerticalSpacer(12),
                              PipelinesList(),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          )
        ],
      ),
    );
  }
}
