import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// Custom theme extension for app-specific colors
@immutable
class AppColors extends ThemeExtension<AppColors> {
  const AppColors({
    required this.primary,
    required this.primaryVariant,
    required this.success,
    required this.warning,
    required this.error,
    required this.info,
    required this.surface,
    required this.cardBackground,
    required this.surfaceVariant,
    required this.textPrimary,
    required this.textSecondary,
    required this.textTertiary,
    required this.border,
    required this.divider,
    required this.shadow,
    required this.inputBackground,
    required this.inputBorder,
    required this.canvasBackground,
    required this.canvasDot,
  });

  final Color primary;
  final Color primaryVariant;
  final Color success;
  final Color warning;
  final Color error;
  final Color info;
  final Color surface;
  final Color cardBackground;
  final Color surfaceVariant;
  final Color textPrimary;
  final Color textSecondary;
  final Color textTertiary;
  final Color border;
  final Color divider;
  final Color shadow;
  final Color inputBackground;
  final Color inputBorder;
  final Color canvasBackground;
  final Color canvasDot;

  // Semantic colors that stay the same in both themes
  static const Color indigo = Color(0xFF6366F1);
  static const Color purple = Color(0xFF8B5CF6);
  static const Color emerald = Color(0xFF10B981);
  static const Color amber = Color(0xFFF59E0B);
  static const Color rose = Color(0xFFEF4444);
  static const Color sky = Color(0xFF0EA5E9);
  static const Color orange = Color(0xFFF97316);

  // Common gradients
  static const LinearGradient primaryGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [indigo, purple],
  );

  /// Creates a subtle gradient for headers/backgrounds
  static LinearGradient subtleGradient(Color color) => LinearGradient(
        colors: [
          color.withValues(alpha: 0.1),
          color.withValues(alpha: 0.05),
        ],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      );

  @override
  AppColors copyWith({
    Color? primary,
    Color? primaryVariant,
    Color? success,
    Color? warning,
    Color? error,
    Color? info,
    Color? surface,
    Color? cardBackground,
    Color? surfaceVariant,
    Color? textPrimary,
    Color? textSecondary,
    Color? textTertiary,
    Color? border,
    Color? divider,
    Color? shadow,
    Color? inputBackground,
    Color? inputBorder,
    Color? canvasBackground,
    Color? canvasDot,
  }) {
    return AppColors(
      primary: primary ?? this.primary,
      primaryVariant: primaryVariant ?? this.primaryVariant,
      success: success ?? this.success,
      warning: warning ?? this.warning,
      error: error ?? this.error,
      info: info ?? this.info,
      surface: surface ?? this.surface,
      cardBackground: cardBackground ?? this.cardBackground,
      surfaceVariant: surfaceVariant ?? this.surfaceVariant,
      textPrimary: textPrimary ?? this.textPrimary,
      textSecondary: textSecondary ?? this.textSecondary,
      textTertiary: textTertiary ?? this.textTertiary,
      border: border ?? this.border,
      divider: divider ?? this.divider,
      shadow: shadow ?? this.shadow,
      inputBackground: inputBackground ?? this.inputBackground,
      inputBorder: inputBorder ?? this.inputBorder,
      canvasBackground: canvasBackground ?? this.canvasBackground,
      canvasDot: canvasDot ?? this.canvasDot,
    );
  }

  @override
  AppColors lerp(ThemeExtension<AppColors>? other, double t) {
    if (other is! AppColors) return this;
    return AppColors(
      primary: Color.lerp(primary, other.primary, t)!,
      primaryVariant: Color.lerp(primaryVariant, other.primaryVariant, t)!,
      success: Color.lerp(success, other.success, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      error: Color.lerp(error, other.error, t)!,
      info: Color.lerp(info, other.info, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      cardBackground: Color.lerp(cardBackground, other.cardBackground, t)!,
      surfaceVariant: Color.lerp(surfaceVariant, other.surfaceVariant, t)!,
      textPrimary: Color.lerp(textPrimary, other.textPrimary, t)!,
      textSecondary: Color.lerp(textSecondary, other.textSecondary, t)!,
      textTertiary: Color.lerp(textTertiary, other.textTertiary, t)!,
      border: Color.lerp(border, other.border, t)!,
      divider: Color.lerp(divider, other.divider, t)!,
      shadow: Color.lerp(shadow, other.shadow, t)!,
      inputBackground: Color.lerp(inputBackground, other.inputBackground, t)!,
      inputBorder: Color.lerp(inputBorder, other.inputBorder, t)!,
      canvasBackground:
          Color.lerp(canvasBackground, other.canvasBackground, t)!,
      canvasDot: Color.lerp(canvasDot, other.canvasDot, t)!,
    );
  }
}

/// Light theme colors
const _lightAppColors = AppColors(
  primary: AppColors.indigo,
  primaryVariant: AppColors.purple,
  success: AppColors.emerald,
  warning: AppColors.amber,
  error: AppColors.rose,
  info: AppColors.sky,
  surface: Color(0xFFFFFFFF),
  cardBackground: Color(0xFFFFFFFF),
  surfaceVariant: Color(0xFFF8FAFC),
  textPrimary: Color(0xFF1E293B),
  textSecondary: Color(0xFF64748B),
  textTertiary: Color(0xFF94A3B8),
  border: Color(0xFFE2E8F0),
  divider: Color(0xFFF1F5F9),
  shadow: Color(0x1A000000),
  inputBackground: Color(0xFFF8FAFC),
  inputBorder: Color(0xFFE2E8F0),
  canvasBackground: Color(0xFFF5F8F9),
  canvasDot: Color(0xFFE2E8F0),
);

/// Dark theme colors
const _darkAppColors = AppColors(
  primary: AppColors.indigo,
  primaryVariant: AppColors.purple,
  success: AppColors.emerald,
  warning: AppColors.amber,
  error: AppColors.rose,
  info: AppColors.sky,
  surface: Color(0xFF1E293B),
  cardBackground: Color(0xFF1E293B),
  surfaceVariant: Color(0xFF0F172A),
  textPrimary: Color(0xFFF1F5F9),
  textSecondary: Color(0xFF94A3B8),
  textTertiary: Color(0xFF64748B),
  border: Color(0xFF334155),
  divider: Color(0xFF1E293B),
  shadow: Color(0x40000000),
  inputBackground: Color(0xFF0F172A),
  inputBorder: Color(0xFF334155),
  canvasBackground: Color(0xFF0F172A),
  canvasDot: Color(0xFF334155),
);

/// App theme configuration
class AppTheme {
  AppTheme._();

  static final _baseTextTheme = GoogleFonts.interTextTheme();

  /// Light theme
  static ThemeData get light {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorScheme: ColorScheme.light(
        primary: AppColors.indigo,
        secondary: AppColors.purple,
        error: AppColors.rose,
        surface: const Color(0xFFFFFFFF),
        onPrimary: Colors.white,
        onSecondary: Colors.white,
        onError: Colors.white,
        onSurface: const Color(0xFF1E293B),
      ),
      scaffoldBackgroundColor: const Color(0xFFF8FAFC),
      cardColor: const Color(0xFFFFFFFF),
      dividerColor: const Color(0xFFE2E8F0),
      textTheme: _baseTextTheme.apply(
        bodyColor: const Color(0xFF1E293B),
        displayColor: const Color(0xFF1E293B),
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: IconThemeData(color: Color(0xFF1E293B)),
        titleTextStyle: TextStyle(
          color: Color(0xFF1E293B),
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      cardTheme: CardThemeData(
        color: const Color(0xFFFFFFFF),
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: Color(0xFFE2E8F0)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFE2E8F0)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.indigo, width: 2),
        ),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.indigo,
          foregroundColor: Colors.white,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.indigo,
          side: const BorderSide(color: AppColors.indigo),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColors.indigo,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        ),
      ),
      iconTheme: const IconThemeData(
        color: Color(0xFF64748B),
        size: 24,
      ),
      tooltipTheme: TooltipThemeData(
        decoration: BoxDecoration(
          color: const Color(0xFF1E293B),
          borderRadius: BorderRadius.circular(8),
        ),
        textStyle: const TextStyle(color: Colors.white, fontSize: 12),
      ),
      extensions: const [_lightAppColors],
    );
  }

  /// Dark theme
  static ThemeData get dark {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: ColorScheme.dark(
        primary: AppColors.indigo,
        secondary: AppColors.purple,
        error: AppColors.rose,
        surface: const Color(0xFF1E293B),
        onPrimary: Colors.white,
        onSecondary: Colors.white,
        onError: Colors.white,
        onSurface: const Color(0xFFF1F5F9),
      ),
      scaffoldBackgroundColor: const Color(0xFF0F172A),
      cardColor: const Color(0xFF1E293B),
      dividerColor: const Color(0xFF334155),
      textTheme: _baseTextTheme.apply(
        bodyColor: const Color(0xFFF1F5F9),
        displayColor: const Color(0xFFF1F5F9),
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: IconThemeData(color: Color(0xFFF1F5F9)),
        titleTextStyle: TextStyle(
          color: Color(0xFFF1F5F9),
          fontSize: 18,
          fontWeight: FontWeight.w600,
        ),
      ),
      cardTheme: CardThemeData(
        color: const Color(0xFF1E293B),
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: Color(0xFF334155)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: const Color(0xFF0F172A),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF334155)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.indigo, width: 2),
        ),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.indigo,
          foregroundColor: Colors.white,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.indigo,
          side: const BorderSide(color: AppColors.indigo),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColors.indigo,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        ),
      ),
      iconTheme: const IconThemeData(
        color: Color(0xFF94A3B8),
        size: 24,
      ),
      tooltipTheme: TooltipThemeData(
        decoration: BoxDecoration(
          color: const Color(0xFF334155),
          borderRadius: BorderRadius.circular(8),
        ),
        textStyle: const TextStyle(color: Colors.white, fontSize: 12),
      ),
      extensions: const [_darkAppColors],
    );
  }
}

/// Extension to easily access AppColors from context
extension AppColorsExtension on BuildContext {
  AppColors get colors => Theme.of(this).extension<AppColors>()!;

  bool get isDarkMode => Theme.of(this).brightness == Brightness.dark;
}
