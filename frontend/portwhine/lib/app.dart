import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/bloc_providers.dart';
import 'package:portwhine/blocs/theme/theme_cubit.dart';
import 'package:portwhine/global/app_scroll_behaviour.dart';
import 'package:portwhine/global/theme.dart';
import 'package:portwhine/router/router.dart';

class PortWhineApp extends StatefulWidget {
  const PortWhineApp({super.key});

  @override
  State<PortWhineApp> createState() => _PortWhineAppState();
}

class _PortWhineAppState extends State<PortWhineApp> {
  final appRouter = AppRouter();

  @override
  Widget build(BuildContext context) {
    return buildBlocProvider();
  }

  Widget buildBlocProvider() {
    return MultiBlocProvider(
      providers: [
        BlocProvider(create: (_) => ThemeCubit()),
        ...BlocProviders.providers,
      ],
      child: buildMaterialApp(),
    );
  }

  Widget buildMaterialApp() {
    return BlocBuilder<ThemeCubit, AppThemeMode>(
      builder: (context, themeMode) {
        return MaterialApp.router(
          title: 'PortWhine',
          debugShowCheckedModeBanner: false,
          scrollBehavior: AppScrollBehavior(),
          theme: AppTheme.light,
          darkTheme: AppTheme.dark,
          themeMode: context.read<ThemeCubit>().themeMode,
          routerConfig: appRouter.config(),
        );
      },
    );
  }
}
