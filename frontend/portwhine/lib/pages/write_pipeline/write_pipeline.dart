import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/pipelines/create_pipeline/create_pipeline_bloc.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/models/pipeline_model.dart';
import 'package:portwhine/widgets/button.dart';
import 'package:portwhine/widgets/spacer.dart';
import 'package:portwhine/widgets/text_field.dart';

class WritePipelineDialog extends StatefulWidget {
  const WritePipelineDialog({this.pipeline, Key? key}) : super(key: key);

  final PipelineModel? pipeline;

  @override
  State<WritePipelineDialog> createState() => _WritePipelineDialogState();
}

class _WritePipelineDialogState extends State<WritePipelineDialog> {
  final nameController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 400,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        color: MyColors.white,
      ),
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          MyTextField(
            hint: 'Name',
            controller: nameController,
          ),
          const VerticalSpacer(12),
          BlocBuilder<CreatePipelineBloc, CreatePipelineState>(
            builder: (context, state) {
              return Button(
                widget.pipeline == null ? 'Create Pipeline' : 'Update Pipeline',
                showLoading: state is CreatePipelineStarted,
                width: double.infinity,
                onTap: () {
                  final name = nameController.text;
                  if (name.isEmpty) return;

                  BlocProvider.of<CreatePipelineBloc>(context).add(
                    CreatePipeline(name),
                  );
                },
              );
            },
          ),
        ],
      ),
    );
  }
}

void showWritePipelineDialog(BuildContext context, {PipelineModel? pipeline}) {
  showDialog(
    context: context,
    builder: (_) {
      return Dialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
        ),
        child: WritePipelineDialog(pipeline: pipeline),
      );
    },
  );
}
