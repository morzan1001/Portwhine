import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:shimmer/shimmer.dart';

class ShimmerEffect extends StatelessWidget {
  const ShimmerEffect({this.length = 3, this.height = 40, super.key});

  final int length;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Shimmer.fromColors(
      baseColor: MyColors.grey,
      highlightColor: MyColors.darkGrey,
      child: Column(
        children: [
          ...List.generate(
            length,
            (i) => Container(
              margin: const EdgeInsets.only(bottom: 12),
              height: height,
              width: double.infinity,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: MyColors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
