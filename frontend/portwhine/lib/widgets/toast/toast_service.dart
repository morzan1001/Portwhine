import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:portwhine/global/colors.dart';
import 'package:portwhine/global/text_style.dart';
import 'package:portwhine/widgets/spacer.dart';

enum ToastType {
  success,
  error,
  info,
  warning;

  Color get color {
    switch (this) {
      case ToastType.success:
        return const Color(0xFF10B981); // Emerald
      case ToastType.error:
        return const Color(0xFFEF4444); // Rose
      case ToastType.info:
        return const Color(0xFF3B82F6); // Blue
      case ToastType.warning:
        return const Color(0xFFF59E0B); // Amber
    }
  }

  IconData get icon {
    switch (this) {
      case ToastType.success:
        return Icons.check_circle_rounded;
      case ToastType.error:
        return Icons.error_rounded;
      case ToastType.info:
        return Icons.info_rounded;
      case ToastType.warning:
        return Icons.warning_rounded;
    }
  }
}

class ToastService {
  static final FToast _fToast = FToast();
  static bool _initialized = false;

  static void init(BuildContext context) {
    if (!_initialized) {
      _fToast.init(context);
      _initialized = true;
    }
  }

  static void show(
    BuildContext context,
    String message, {
    ToastType type = ToastType.info,
    Duration duration = const Duration(seconds: 4),
  }) {
    init(context);

    _fToast.showToast(
      child: _ToastWidget(message: message, type: type),
      gravity: ToastGravity.TOP,
      toastDuration: duration,
      fadeDuration: const Duration(milliseconds: 200),
    );
  }

  static void success(BuildContext context, String message) =>
      show(context, message, type: ToastType.success);

  static void error(BuildContext context, String message) =>
      show(context, message, type: ToastType.error);

  static void info(BuildContext context, String message) =>
      show(context, message, type: ToastType.info);

  static void warning(BuildContext context, String message) =>
      show(context, message, type: ToastType.warning);
}

class _ToastWidget extends StatelessWidget {
  const _ToastWidget({required this.message, required this.type});

  final String message;
  final ToastType type;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(top: 24),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: type.color.withValues(alpha: 0.15),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
        border: Border.all(color: type.color.withValues(alpha: 0.2), width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: type.color.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(type.icon, color: type.color, size: 20),
          ),
          const HorizontalSpacer(12),
          Flexible(
            child: Text(
              message,
              style: style(
                color: Colors.black87,
                weight: FontWeight.w500,
                size: 14,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}
