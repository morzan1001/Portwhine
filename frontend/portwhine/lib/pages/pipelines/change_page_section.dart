import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/pipeline_page/pipeline_page_cubit.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/widgets/icon_button.dart';
import 'package:portwhine/widgets/text.dart';

class ChangePageSection extends StatelessWidget {
  const ChangePageSection({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<PipelinePageCubit, int>(
      builder: (context, state) {
        return Row(
          mainAxisAlignment: MainAxisAlignment.center,
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
                child: Heading('$state'),
              ),
            ),
            MyIconButton(
              Icons.arrow_right,
              buttonColor: MyColors.white,
              onTap: () {
                BlocProvider.of<PipelinePageCubit>(context).nextPage();
              },
            ),
          ],
        );
      },
    );
  }
}
