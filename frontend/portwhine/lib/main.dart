import 'package:flutter/material.dart';
import 'package:flutter_web_plugins/flutter_web_plugins.dart';
import 'package:portwhine/api/api.dart';
import 'package:portwhine/app.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  usePathUrlStrategy();
  await initApi();
  runApp(const PortWhineApp());
}
