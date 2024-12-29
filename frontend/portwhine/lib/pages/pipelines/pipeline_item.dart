import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/router/router.dart';
import 'package:portwhine/widgets/spacer.dart';

class PipelineItem extends StatelessWidget {
  const PipelineItem(this.model, {super.key});

  final PipelineModel model;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () {
        AutoRouter.of(context).navigate(
          PipelineDetailsRoute(id: '123', model: model),
        );
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          color: CustomColors.white,
        ),
        child: Row(
          children: [
            // name
            Expanded(
              flex: 2,
              child: Text(
                model.name,
                style: style(
                  color: CustomColors.textDark,
                  weight: FontWeight.w600,
                  size: 14,
                ),
              ),
            ),
            const HorizontalSpacer(12),

            // // run
            // Expanded(
            //   child: Column(
            //     crossAxisAlignment: CrossAxisAlignment.start,
            //     children: [
            //       Text(
            //         'RUN',
            //         style: style(
            //           color: CustomColors.textLight,
            //           weight: FontWeight.w400,
            //           size: 12,
            //         ),
            //       ),
            //       Text(
            //         model.runningTime.toString(),
            //         style: style(
            //           color: CustomColors.textDark,
            //           weight: FontWeight.w600,
            //           size: 14,
            //         ),
            //       ),
            //     ],
            //   ),
            // ),
            // const HorizontalSpacer(12),

            // status
            Expanded(
              flex: 2,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'STATUS',
                    style: style(
                      color: CustomColors.textLight,
                      weight: FontWeight.w400,
                      size: 12,
                    ),
                  ),
                  Row(
                    children: [
                      Container(
                        height: 8,
                        width: 18,
                        color: CustomColors.green,
                      ),
                      const HorizontalSpacer(2),
                      Container(
                        height: 8,
                        width: 18,
                        color: CustomColors.green,
                      ),
                      const HorizontalSpacer(2),
                      Container(
                        height: 8,
                        width: 18,
                        color: CustomColors.yellow,
                      ),
                    ],
                  ),
                  Text(
                    model.status ?? '',
                    style: style(
                      color: CustomColors.textDark,
                      weight: FontWeight.w600,
                      size: 12,
                    ),
                  ),
                ],
              ),
            ),
            const HorizontalSpacer(12),

            // errors
            Expanded(
              flex: 1,
              child: Row(
                children: [
                  if (model.errors > 0)
                    const Icon(
                      Icons.report_outlined,
                      color: CustomColors.error,
                    ),
                  if (model.errors > 0) const HorizontalSpacer(4),
                  Text(
                    '${model.errors > 0 ? model.errors : 'No'} errors',
                    style: style(
                      color: model.errors > 0
                          ? CustomColors.error
                          : CustomColors.green,
                      weight: FontWeight.w500,
                      size: 12,
                    ),
                  ),
                ],
              ),
            ),
            const Spacer(),
            InkWell(
              onTap: () {},
              child: Container(
                padding: const EdgeInsets.symmetric(
                  vertical: 6,
                  horizontal: 10,
                ),
                decoration: BoxDecoration(
                  border: Border.all(color: CustomColors.grey),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.play_arrow,
                  size: 20,
                ),
              ),
            ),
            const HorizontalSpacer(8),
            // InkWell(
            //   onTap: () {},
            //   child: Container(
            //     padding: const EdgeInsets.symmetric(
            //       vertical: 6,
            //       horizontal: 10,
            //     ),
            //     decoration: BoxDecoration(
            //       border: Border.all(color: CustomColors.grey),
            //       borderRadius: BorderRadius.circular(8),
            //     ),
            //     child: const Icon(
            //       Icons.pause,
            //       size: 20,
            //     ),
            //   ),
            // ),
            // const HorizontalSpacer(8),
            // InkWell(
            //   onTap: () {},
            //   child: Container(
            //     padding: const EdgeInsets.symmetric(
            //       vertical: 6,
            //       horizontal: 10,
            //     ),
            //     decoration: BoxDecoration(
            //       border: Border.all(color: CustomColors.grey),
            //       borderRadius: BorderRadius.circular(8),
            //     ),
            //     child: const Icon(
            //       Icons.refresh,
            //       size: 20,
            //     ),
            //   ),
            // ),
            // const HorizontalSpacer(16),
            InkWell(
              onTap: () {},
              child: Container(
                padding: const EdgeInsets.symmetric(
                  vertical: 6,
                  horizontal: 10,
                ),
                decoration: BoxDecoration(
                  border: Border.all(color: CustomColors.red),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.delete,
                  size: 20,
                  color: CustomColors.red,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
