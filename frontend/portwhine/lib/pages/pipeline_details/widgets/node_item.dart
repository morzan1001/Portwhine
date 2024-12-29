import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';

class ResultNodeItem extends StatelessWidget {
  const ResultNodeItem({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        border: Border.all(color: CustomColors.grey, width: 0.5),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Theme(
        data: ThemeData(
          dividerColor: Colors.transparent,
        ),
        child: ExpansionTile(
          tilePadding: const EdgeInsets.symmetric(
            horizontal: 20,
            vertical: 4,
          ),
          title: Text(
            'Domain',
            style: style(
              color: CustomColors.textDark,
              weight: FontWeight.w600,
            ),
          ),
          childrenPadding: EdgeInsets.zero,
          children: [
            Container(
              color: CustomColors.greyLighter,
              padding: const EdgeInsets.symmetric(
                horizontal: 20,
                vertical: 8,
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      'OUTPUTS',
                      style: style(color: CustomColors.textLight),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      'NAME',
                      style: style(color: CustomColors.textLight),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      'TYPE',
                      style: style(color: CustomColors.textLight),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      'VALUE',
                      style: style(color: CustomColors.textLight),
                    ),
                  ),
                ],
              ),
            ),
            ListView.separated(
              separatorBuilder: (a, b) => const Divider(
                height: 0.5,
                thickness: 0.5,
                color: CustomColors.grey,
              ),
              itemCount: 4,
              shrinkWrap: true,
              itemBuilder: (context, index) {
                return Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 8,
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          'DNS Ports',
                          style: style(color: CustomColors.black),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          'Open DNS Port',
                          style: style(color: CustomColors.black),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          'UDP Port',
                          style: style(color: CustomColors.black),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          '53',
                          style: style(color: CustomColors.black),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
