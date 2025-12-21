import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/single_pipeline/workers_list/workers_list_bloc.dart';
import 'package:portwhine/pages/pipeline_details/canvas/draggable_node_item.dart';
import 'package:portwhine/pages/pipeline_details/canvas/node_list_container.dart';
import 'package:portwhine/widgets/shimmer.dart';
import 'package:portwhine/widgets/spacer.dart';

class WorkersList extends StatelessWidget {
  const WorkersList({super.key});

  @override
  Widget build(BuildContext context) {
    return NodeListContainer(
      title: 'Workers',
      child: BlocBuilder<WorkersListBloc, WorkersListState>(
        builder: (context, state) {
          if (state is WorkersListLoading) {
            return const ShimmerEffect();
          }

          if (state is WorkersListFailed) {
            return Text(state.error);
          }

          if (state is WorkersListLoaded) {
            final workers = state.workers;

            return ListView.separated(
              separatorBuilder: (a, b) => const VerticalSpacer(12),
              itemCount: workers.length,
              shrinkWrap: true,
              itemBuilder: (_, i) {
                return DraggableNodeItem(workers[i]);
              },
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }
}
