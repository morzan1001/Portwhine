import 'package:flutter/material.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/pages/pipelines/pipelines.dart';
import 'package:portwhine/router/router.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final _appRouter = AppRouter();

    return MaterialApp.router(
      title: 'Portwhine',
      theme: ThemeData(
        primaryColor: CustomColors.prime,
        colorScheme: ColorScheme.fromSwatch().copyWith(
          secondary: CustomColors.sec,
          error: CustomColors.error,
        ),
        textTheme: TextTheme(
          headline1: TextStyle(
            fontSize: 32,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          headline2: TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          headline3: TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          headline4: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          headline5: TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          headline6: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: CustomColors.textDark,
          ),
          subtitle1: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w500,
            color: CustomColors.textLight,
          ),
          subtitle2: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: CustomColors.textLight,
          ),
          bodyText1: TextStyle(
            fontSize: 16,
            color: CustomColors.textDark,
          ),
          bodyText2: TextStyle(
            fontSize: 14,
            color: CustomColors.textDark,
          ),
          button: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: CustomColors.white,
          ),
          caption: TextStyle(
            fontSize: 12,
            color: CustomColors.textLight,
          ),
          overline: TextStyle(
            fontSize: 10,
            color: CustomColors.textLight,
          ),
        ),
      ),
      routerDelegate: _appRouter.delegate(),
      routeInformationParser: _appRouter.defaultRouteParser(),
    );
  }
}
