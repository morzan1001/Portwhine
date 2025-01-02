import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:portwhine/blocs/bloc_providers.dart';
import 'package:portwhine/global/app_scroll_behaviour.dart';
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
      providers: BlocProviders.providers,
      child: buildMaterialApp(),
    );
  }

  buildMaterialApp() {
    return MaterialApp.router(
      title: 'PortWhine',
      debugShowCheckedModeBanner: false,
      scrollBehavior: AppScrollBehavior(),
      theme: ThemeData(
        fontFamily: GoogleFonts.inter().fontFamily,
      ),
      routerConfig: appRouter.config(),
    );
  }
}
