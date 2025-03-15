import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/triggers_list/triggers_list_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/shimmer.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/text.dart';

class TriggersList extends StatelessWidget {
  const TriggersList({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 280,
      padding: const EdgeInsets.symmetric(
        horizontal: 20,
        vertical: 12,
      ),
      decoration: BoxDecoration(
        color: MyColors.darkGrey.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: MyColors.darkGrey, width: 0.5),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Heading('Triggers'),
          const VerticalSpacer(12),
          Expanded(
            child: BlocBuilder<TriggersListBloc, TriggersListState>(
              builder: (context, state) {
                if (state is TriggersListLoading) {
                  return const ShimmerEffect();
                }

                if (state is TriggersListFailed) {
                  return Text(state.error);
                }

                if (state is TriggersListLoaded) {
                  final triggers = state.triggers;

                  return ListView.separated(
                    separatorBuilder: (a, b) => const VerticalSpacer(12),
                    itemCount: triggers.length,
                    shrinkWrap: true,
                    itemBuilder: (_, i) {
                      return SelectionItem(triggers[i]);
                    },
                  );
                }

                return const SizedBox.shrink();
              },
            ),
          ),
        ],
      ),
    );
  }
}

class SelectionItem extends StatelessWidget {
  const SelectionItem(this.name, {super.key});

  final String name;

  @override
  Widget build(BuildContext context) {
    return Material(
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: MyColors.white,
          borderRadius: BorderRadius.circular(6),
        ),
        child: Row(
          children: [
            const Icon(
              Icons.drag_indicator,
              color: MyColors.black,
              size: 20,
            ),
            const HorizontalSpacer(4),
            Text(
              name,
              style: style(
                weight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
