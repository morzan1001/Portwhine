import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:portwhine/blocs/theme/theme_cubit.dart';
import 'package:portwhine/global/theme.dart';

/// A button to toggle between light and dark theme
class ThemeToggleButton extends StatefulWidget {
  const ThemeToggleButton({this.size = 40, super.key});

  final double size;

  @override
  State<ThemeToggleButton> createState() => _ThemeToggleButtonState();
}

class _ThemeToggleButtonState extends State<ThemeToggleButton> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final isDark = context.isDarkMode;

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () => context.read<ThemeCubit>().toggleTheme(),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          width: widget.size,
          height: widget.size,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            color: _isHovered
                ? colors.primary.withValues(alpha: 0.1)
                : Colors.transparent,
          ),
          child: Center(
            child: AnimatedSwitcher(
              duration: const Duration(milliseconds: 300),
              transitionBuilder: (child, animation) {
                return RotationTransition(
                  turns: animation,
                  child: FadeTransition(opacity: animation, child: child),
                );
              },
              child: Icon(
                isDark ? Icons.light_mode_rounded : Icons.dark_mode_rounded,
                key: ValueKey(isDark),
                color: _isHovered ? colors.primary : colors.textSecondary,
                size: widget.size * 0.5,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// A segmented control for theme selection
class ThemeSegmentedControl extends StatelessWidget {
  const ThemeSegmentedControl({super.key});

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;

    return BlocBuilder<ThemeCubit, AppThemeMode>(
      builder: (context, mode) {
        return Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: colors.surfaceVariant,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              _ThemeOption(
                icon: Icons.light_mode_rounded,
                label: 'Light',
                isSelected: mode == AppThemeMode.light,
                onTap: () =>
                    context.read<ThemeCubit>().setTheme(AppThemeMode.light),
              ),
              _ThemeOption(
                icon: Icons.settings_rounded,
                label: 'System',
                isSelected: mode == AppThemeMode.system,
                onTap: () =>
                    context.read<ThemeCubit>().setTheme(AppThemeMode.system),
              ),
              _ThemeOption(
                icon: Icons.dark_mode_rounded,
                label: 'Dark',
                isSelected: mode == AppThemeMode.dark,
                onTap: () =>
                    context.read<ThemeCubit>().setTheme(AppThemeMode.dark),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _ThemeOption extends StatelessWidget {
  const _ThemeOption({
    required this.icon,
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;

    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? colors.cardBackground : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: colors.shadow,
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 18,
              color: isSelected ? colors.primary : colors.textTertiary,
            ),
            const SizedBox(width: 8),
            Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                color: isSelected ? colors.textPrimary : colors.textTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
