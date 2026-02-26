// ignore_for_file: prefer_const_constructors

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

// ═══════════════════════════════════════════════════════════════════════════
// COLOR SYSTEM
// ═══════════════════════════════════════════════════════════════════════════

class AppColors {
  AppColors._();

  // Primary Palette (Purple)
  static const Color primary50 = Color(0xFFF5F3FF);
  static const Color primary100 = Color(0xFFEDE9FE);
  static const Color primary200 = Color(0xFFDDD6FE);
  static const Color primary300 = Color(0xFFC4B5FD);
  static const Color primary400 = Color(0xFFA78BFA);
  static const Color primary500 = Color(0xFF8B5CF6);
  static const Color primary600 = Color(0xFF7C3AED);
  static const Color primary700 = Color(0xFF6D28D9);
  static const Color primary800 = Color(0xFF5B21B6);
  static const Color primary900 = Color(0xFF4C1D95);

  // Accent
  static const Color warning = Color(0xFFD97706);
  static const Color error = Color(0xFFDC2626);
  static const Color success = Color(0xFF059669);
  static const Color info = Color(0xFF0284C7);

  // Light Theme
  static const Color lightBgDefault = Color(0xFFFFFFFF);
  static const Color lightBgSecondary = Color(0xFFF6F8FA);
  static const Color lightBgTertiary = Color(0xFFEAEEF2);
  static const Color lightTextPrimary = Color(0xFF24292F);
  static const Color lightTextSecondary = Color(0xFF57606A);
  static const Color lightTextTertiary = Color(0xFF6E7781);
  static const Color lightTextMuted = Color(0xFF8C959F);
  static const Color lightBorderPrimary = Color(0xFFD0D7DE);
  static const Color lightBorderSecondary = Color(0xFFEAEEF2);

  // Dark Theme
  static const Color darkBgDefault = Color(0xFF0D1117);
  static const Color darkBgSecondary = Color(0xFF010409);
  static const Color darkBgTertiary = Color(0xFF161B22);
  static const Color darkTextPrimary = Color(0xFFC9D1D9);
  static const Color darkTextSecondary = Color(0xFF8B949E);
  static const Color darkTextTertiary = Color(0xFF484F58);
  static const Color darkTextMuted = Color(0xFF30363D);
  static const Color darkBorderPrimary = Color(0xFF30363D);
  static const Color darkBorderSecondary = Color(0xFF21262D);
}

// ═══════════════════════════════════════════════════════════════════════════
// THEME-AWARE EXTENSION
// ═══════════════════════════════════════════════════════════════════════════

class AppThemeColors {
  final BuildContext _context;
  AppThemeColors(this._context);

  bool get _isDark => Theme.of(_context).brightness == Brightness.dark;

  Color get bgDefault => _isDark ? AppColors.darkBgDefault : AppColors.lightBgDefault;
  Color get bgSecondary => _isDark ? AppColors.darkBgSecondary : AppColors.lightBgSecondary;
  Color get bgTertiary => _isDark ? AppColors.darkBgTertiary : AppColors.lightBgTertiary;
  Color get textPrimary => _isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
  Color get textSecondary => _isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
  Color get textTertiary => _isDark ? AppColors.darkTextTertiary : AppColors.lightTextTertiary;
  Color get textMuted => _isDark ? AppColors.darkTextMuted : AppColors.lightTextMuted;
  Color get borderPrimary => _isDark ? AppColors.darkBorderPrimary : AppColors.lightBorderPrimary;
  Color get borderSecondary => _isDark ? AppColors.darkBorderSecondary : AppColors.lightBorderSecondary;

  BoxDecoration get cardDecoration => BoxDecoration(
        color: bgDefault,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: borderSecondary),
        boxShadow: _isDark
            ? []
            : [BoxShadow(color: Colors.grey.withValues(alpha: 0.08), blurRadius: 20, offset: Offset(0, 5))],
      );
}

extension ThemeColorsExtension on BuildContext {
  AppThemeColors get colors => AppThemeColors(this);
  bool get isDark => Theme.of(this).brightness == Brightness.dark;
}

// ═══════════════════════════════════════════════════════════════════════════
// THEME DATA
// ═══════════════════════════════════════════════════════════════════════════

final ThemeData lightTheme = ThemeData(
  brightness: Brightness.light,
  scaffoldBackgroundColor: AppColors.lightBgSecondary,
  primaryColor: AppColors.primary500,
  colorScheme: ColorScheme.light(
    primary: AppColors.primary500,
    secondary: AppColors.primary400,
    surface: AppColors.lightBgDefault,
    error: AppColors.error,
  ),
  appBarTheme: AppBarTheme(
    backgroundColor: AppColors.lightBgDefault,
    foregroundColor: AppColors.lightTextPrimary,
    elevation: 0,
    systemOverlayStyle: SystemUiOverlayStyle.dark,
  ),
  bottomNavigationBarTheme: BottomNavigationBarThemeData(
    backgroundColor: AppColors.lightBgDefault,
    selectedItemColor: AppColors.primary500,
    unselectedItemColor: AppColors.lightTextTertiary,
  ),
  useMaterial3: true,
);

final ThemeData darkTheme = ThemeData(
  brightness: Brightness.dark,
  scaffoldBackgroundColor: AppColors.darkBgSecondary,
  primaryColor: AppColors.primary500,
  colorScheme: ColorScheme.dark(
    primary: AppColors.primary500,
    secondary: AppColors.primary400,
    surface: AppColors.darkBgDefault,
    error: AppColors.error,
  ),
  appBarTheme: AppBarTheme(
    backgroundColor: AppColors.darkBgDefault,
    foregroundColor: AppColors.darkTextPrimary,
    elevation: 0,
    systemOverlayStyle: SystemUiOverlayStyle.light,
  ),
  bottomNavigationBarTheme: BottomNavigationBarThemeData(
    backgroundColor: AppColors.darkBgDefault,
    selectedItemColor: AppColors.primary500,
    unselectedItemColor: AppColors.darkTextTertiary,
  ),
  useMaterial3: true,
);
