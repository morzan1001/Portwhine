import 'package:flutter/material.dart';
import 'package:portwhine/global/theme.dart';

/// Centralized helper for node-related styling and logic.
/// Avoids repeating node type detection and color/icon mapping.
class NodeHelper {
  NodeHelper._();

  /// Returns the color for a node based on its name.
  static Color getNodeColor(String nodeName) {
    final name = nodeName.toLowerCase();

    if (name.contains('trigger') ||
        name.contains('certstream') ||
        name.contains('ip')) {
      return AppColors.emerald;
    } else if (name.contains('nmap') || name.contains('scan')) {
      return AppColors.amber;
    } else if (name.contains('screenshot') || name.contains('web')) {
      return AppColors.indigo;
    } else if (name.contains('ssl') ||
        name.contains('security') ||
        name.contains('testssl')) {
      return AppColors.rose;
    } else if (name.contains('humble')) {
      return AppColors.purple;
    } else if (name.contains('resolver') || name.contains('dns')) {
      return AppColors.sky;
    } else if (name.contains('ffuf') || name.contains('fuzz')) {
      return AppColors.orange;
    }

    return AppColors.purple; // Default
  }

  /// Returns the icon for a node based on its name.
  static IconData getNodeIcon(String nodeName) {
    final name = nodeName.toLowerCase();

    if (name.contains('trigger') || name.contains('certstream')) {
      return Icons.bolt_rounded;
    } else if (name.contains('ip')) {
      return Icons.dns_rounded;
    } else if (name.contains('nmap') || name.contains('scan')) {
      return Icons.radar_rounded;
    } else if (name.contains('screenshot')) {
      return Icons.camera_alt_rounded;
    } else if (name.contains('web') || name.contains('webapp')) {
      return Icons.language_rounded;
    } else if (name.contains('ssl') || name.contains('testssl')) {
      return Icons.security_rounded;
    } else if (name.contains('humble')) {
      return Icons.shield_rounded;
    } else if (name.contains('resolver') || name.contains('dns')) {
      return Icons.search_rounded;
    } else if (name.contains('ffuf') || name.contains('fuzz')) {
      return Icons.find_replace_rounded;
    }

    return Icons.extension_rounded; // Default
  }

  /// Input connector color (green)
  static const Color inputColor = AppColors.emerald;

  /// Output connector color (indigo)
  static const Color outputColor = AppColors.indigo;
}

/// Status helper for consistent status colors.
class StatusHelper {
  StatusHelper._();

  static Color getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'running':
        return AppColors.emerald;
      case 'error':
        return AppColors.rose;
      case 'stopped':
        return AppColors.amber;
      default:
        return const Color(0xFF64748B); // textSecondary
    }
  }

  static bool isRunning(String status) => status.toLowerCase() == 'running';
  static bool isError(String status) => status.toLowerCase() == 'error';
  static bool isStopped(String status) => status.toLowerCase() == 'stopped';
}
