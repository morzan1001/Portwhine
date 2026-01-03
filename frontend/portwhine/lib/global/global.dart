import 'package:flutter/material.dart';

double width(BuildContext context) {
  return MediaQuery.of(context).size.width;
}

double height(BuildContext context) {
  return MediaQuery.of(context).size.height;
}

Future<dynamic> navigate(BuildContext context, String to, {Object? arguments}) {
  return Navigator.of(context).pushNamed(to, arguments: arguments);
}

Future<dynamic> navigateRemove(BuildContext context, String to,
    {Object? arguments}) {
  return Navigator.of(context).pushNamedAndRemoveUntil(
    to,
    (_) => false,
    arguments: arguments,
  );
}

void pop(BuildContext context, [Object? value]) {
  Navigator.of(context).pop(value);
}

Future<dynamic> nav(BuildContext context, Widget to) async {
  return await Navigator.of(context).push(
    MaterialPageRoute(builder: (_) => to),
  );
}
