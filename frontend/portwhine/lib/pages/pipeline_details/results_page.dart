import 'package:auto_route/auto_route.dart';
import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

@RoutePage()
class ResultsPage extends StatelessWidget {
  const ResultsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Pipeline Results', style: style(color: MyColors.black)),
        backgroundColor: MyColors.white,
        elevation: 0,
        iconTheme: const IconThemeData(color: MyColors.black),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.analytics_outlined,
                size: 64, color: MyColors.grey),
            const VerticalSpacer(16),
            Text(
              'Results visualization coming soon',
              style: style(color: MyColors.grey, size: 18),
            ),
          ],
        ),
      ),
    );
  }
}
