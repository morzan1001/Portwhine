import 'package:flutter/material.dart';

double width(BuildContext context) {
  return MediaQuery.of(context).size.width;
}

double height(BuildContext context) {
  return MediaQuery.of(context).size.height;
}

Future navigate(context, String to, {arguments}) {
  return Navigator.of(context).pushNamed(to, arguments: arguments);
}

Future navigateRemove(context, String to, {arguments}) {
  return Navigator.of(context).pushNamedAndRemoveUntil(
    to,
    (_) => false,
    arguments: arguments,
  );
}

void pop(context, [Object? value]) {
  Navigator.of(context).pop(value);
}

Future nav(context, Widget to) async {
  return await Navigator.of(context).push(
    MaterialPageRoute(builder: (_) => to),
  );
}
