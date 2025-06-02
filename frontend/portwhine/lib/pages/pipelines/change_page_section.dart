import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_size_cubit.dart';
import 'package:portwhine/blocs/pipelines/pipelines_list/pipelines_list_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/icon_button.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/text.dart';

class PaginationSection extends StatelessWidget {
  const PaginationSection({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinesListBloc, PipelinesListState>(
      builder: (context, pipelinesState) {
        final isLoaded = pipelinesState is PipelinesListLoaded;

        return AbsorbPointer(
          absorbing: !isLoaded,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const HorizontalSpacer(120),
              PageChangeWidget(isLoaded: isLoaded),
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
    return BlocBuilder<PipelinesListBloc, PipelinesListState>(
      builder: (context, pipelinesState) {
        return BlocBuilder<PipelineSizeCubit, int>(
          builder: (context, size) {
            return BlocBuilder<PipelinePageCubit, int>(
              builder: (context, page) {
                return Row(
                  children: [
                    if (page != 1)
                      MyIconButton(
                        Icons.arrow_left,
                        buttonColor: MyColors.white,
                        onTap: () {
                          BlocProvider.of<PipelinePageCubit>(context)
                              .previousPage();
                        },
                      ),
                    const HorizontalSpacer(12),
                    Center(
                      child: SmallText('Page $page'),
                    ),
                    const HorizontalSpacer(12),
                    if (pipelinesState is PipelinesListLoaded &&
                        pipelinesState.pipelines.length >= size)
                      MyIconButton(
                        Icons.arrow_right,
                        buttonColor: MyColors.white,
                        onTap: () {
                          BlocProvider.of<PipelinePageCubit>(context)
                              .nextPage();
                        },
                      ),
                  ],
                );
              },
            );
          },
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
        return Container(
          width: 120,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            color: MyColors.white,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Center(
            child: DropdownButton<int>(
              focusColor: Colors.transparent,
              underline: const SizedBox.shrink(),
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
            ),
          ),
        );
      },
    );
  }
}
