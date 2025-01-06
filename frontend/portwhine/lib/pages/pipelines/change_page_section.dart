import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/get_all_pipelines/get_all_pipelines_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_size_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/icon_button.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/text.dart';

class PaginationSection extends StatelessWidget {
  const PaginationSection({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<GetAllPipelinesBloc, GetAllPipelinesState>(
      builder: (context, pipelinesState) {
        final isLoaded = pipelinesState is GetAllPipelinesLoaded;

        return AbsorbPointer(
          absorbing: !isLoaded,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              PageChangeWidget(isLoaded: isLoaded),
              const HorizontalSpacer(24),
              const PageSizeDropdownWidget(),
            ],
          ),
        );
      },
    );
  }
}

class PageChangeWidget extends StatelessWidget {
  final bool isLoaded;

  const PageChangeWidget({super.key, required this.isLoaded});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinePageCubit, int>(
      builder: (context, page) {
        return Row(
          children: [
            MyIconButton(
              Icons.arrow_left,
              buttonColor: MyColors.white,
              onTap: () {
                BlocProvider.of<PipelinePageCubit>(context).previousPage();
              },
            ),
            SizedBox(
              width: 40,
              child: Center(
                child: Heading('$page'),
              ),
            ),
            MyIconButton(
              Icons.arrow_right,
              buttonColor: MyColors.white,
              onTap: () {
                if (!isLoaded) return;

                final size = BlocProvider.of<PipelineSizeCubit>(context).state;
                final pipelines = (BlocProvider.of<GetAllPipelinesBloc>(context)
                        .state as GetAllPipelinesLoaded)
                    .pipelines;

                if (pipelines.length >= size) {
                  BlocProvider.of<PipelinePageCubit>(context).nextPage();
                }
              },
            ),
          ],
        );
      },
    );
  }
}

class PageSizeDropdownWidget extends StatelessWidget {
  const PageSizeDropdownWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelineSizeCubit, int>(
      builder: (context, size) {
        return DropdownButton<int>(
          focusColor: Colors.transparent,
          value: size,
          items: [10, 20, 50]
              .map(
                (e) => DropdownMenuItem(
                  value: e,
                  child: Text('$e Items', style: style()),
                ),
              )
              .toList(),
          onChanged: (value) {
            if (value != null) {
              BlocProvider.of<PipelineSizeCubit>(context).setSize(value);
            }
          },
        );
      },
    );
  }
}
