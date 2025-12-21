import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/triggers_list/triggers_list_bloc.dart';
import 'package:portwhine/pages/pipeline_details/canvas/draggable_node_item.dart';
import 'package:portwhine/pages/pipeline_details/canvas/node_list_container.dart';
import 'package:portwhine/widgets/shimmer.dart';
import 'package:portwhine/widgets/spacer.dart';

class TriggersList extends StatelessWidget {
  const TriggersList({super.key});

  @override
  Widget build(BuildContext context) {
    return NodeListContainer(
      title: 'Triggers',
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
                return DraggableNodeItem(triggers[i]);
              },
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }
}
