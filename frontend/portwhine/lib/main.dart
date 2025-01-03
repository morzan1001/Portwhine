import 'package:flutter/material.dart';
import 'package:flutter_web_plugins/flutter_web_plugins.dart';
import 'package:http/http.dart' as http;
import 'package:portwhine/api/api.dart';
import 'package:portwhine/app.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  usePathUrlStrategy();
  await Api.init();
  runApp(const PortWhineApp());
}

void test() async {
  const u = 'https://37.27.179.252:8000/api/v1';
  final list = await http.get(
    Uri.parse('$u/pipelines'),
  );
  final delete = await http.delete(
    Uri.parse('$u/pipeline/205d619e-3713-4882-89d0-cc65e31482b3'),
  );
  debugPrint(list.body);
  debugPrint(delete.body);
}
