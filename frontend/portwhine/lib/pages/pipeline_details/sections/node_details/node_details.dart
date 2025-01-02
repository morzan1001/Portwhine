import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/models/node_model.dart';
import 'package:portwhine/widgets/spacer.dart';

class NodeDetails extends StatelessWidget {
  const NodeDetails(this.model, {super.key});

  final NodeModel model;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 500,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: MyColors.darkGrey.withOpacity(0.1),
        borderRadius: const BorderRadius.vertical(
          top: Radius.circular(16),
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            model.name,
            style: style(
              color: MyColors.black,
              size: 20,
              weight: FontWeight.w600,
            ),
          ),
          const VerticalSpacer(16),
          Text(
            'Settings',
            style: style(
              color: MyColors.black,
              size: 18,
              weight: FontWeight.w500,
            ),
          ),
          const VerticalSpacer(16),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Inputs',
                      style: style(
                        color: MyColors.black,
                        size: 18,
                        weight: FontWeight.w500,
                      ),
                    ),
                    ...List.generate(
                      model.inputs.length,
                      (index) {
                        return Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          margin: const EdgeInsets.only(top: 12),
                          decoration: BoxDecoration(
                            color: MyColors.black.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Center(
                            child: Text(
                              model.inputs.entries.toList()[index].key,
                              style: style(),
                            ),
                          ),
                        );
                      },
                    )
                  ],
                ),
              ),
              const HorizontalSpacer(16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Outputs',
                      style: style(
                        color: MyColors.black,
                        size: 18,
                        weight: FontWeight.w500,
                      ),
                    ),
                    ...List.generate(
                      model.outputs.length,
                      (index) {
                        return Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          margin: const EdgeInsets.only(top: 12),
                          decoration: BoxDecoration(
                            color: MyColors.black.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Center(
                            child: Text(
                              model.outputs.entries.toList()[index].key,
                              style: style(),
                            ),
                          ),
                        );
                      },
                    )
                  ],
                ),
              ),
            ],
          )
        ],
      ),
    );
  }
}

Future showNodeDetailsDialog(BuildContext context, NodeModel model) async {
  return await showDialog(
    context: context,
    barrierColor: Colors.transparent,
    builder: (context) {
      return Dialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        child: NodeDetails(model),
      );
    },
  );
}
