import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/pages/workflows/sections/list/list.dart';
import 'package:portwhine/pages/workflows/sections/overview/overview.dart';
import 'package:portwhine/widgets/spacer.dart';

@RoutePage()
class WorkflowsPage extends StatelessWidget {
  const WorkflowsPage({super.key});

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
                              WorkflowsOverview(),
                              VerticalSpacer(12),
                              WorkflowsList(),
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
