import '../domain/auth_failure.dart';
import '../domain/auth_token_data.dart';
import '../domain/user_profile.dart';
import 'auth_api_client.dart';
import 'token_storage.dart';

class AuthRepository {
  const AuthRepository({
    required AuthApi api,
    required TokenStorage tokenStorage,
  }) : _api = api,
       _tokenStorage = tokenStorage;

  final AuthApi _api;
  final TokenStorage _tokenStorage;

  Future<UserProfile> login({
    required String email,
    required String password,
  }) async {
    final tokens = await _api.login(email: email, password: password);
    final user = _validateEmployeeUser(
      await _api.me(accessToken: tokens.accessToken),
    );
    await _tokenStorage.saveTokens(tokens);
    return user;
  }

  Future<UserProfile?> restoreSession() async {
    final accessToken = await _tokenStorage.readAccessToken();
    if (accessToken == null || accessToken.isEmpty) {
      await _tokenStorage.clearTokens();
      return null;
    }

    try {
      final user = await _api.me(accessToken: accessToken);
      return _validateEmployeeUser(user);
    } on AuthFailure catch (error) {
      if (isTransientAuthFailure(error)) {
        rethrow;
      }
      if (error.kind != AuthFailureKind.sessionExpired) {
        await _tokenStorage.clearTokens();
        return null;
      }
    }

    try {
      final refreshed = await refreshSession();
      if (refreshed == null) {
        return null;
      }
      return getCurrentUser();
    } on AuthFailure catch (error) {
      if (isAuthoritativeSessionFailure(error)) {
        await _tokenStorage.clearTokens();
        return null;
      }
      rethrow;
    }
  }

  Future<UserProfile?> restoreSessionOrNull() async {
    try {
      return await restoreSession();
    } on AuthFailure {
      return null;
    }
  }

  Future<UserProfile> getCurrentUser() async {
    final accessToken = await _tokenStorage.readAccessToken();
    if (accessToken == null || accessToken.isEmpty) {
      throw const AuthFailure(
        AuthFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    }

    final user = await _api.me(accessToken: accessToken);
    return _validateEmployeeUser(user);
  }

  Future<AuthTokenData?> refreshSession() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      await _tokenStorage.clearTokens();
      return null;
    }

    try {
      final tokens = await _api.refresh(refreshToken: refreshToken);
      _validateEmployeeUser(tokens.user);
      await _tokenStorage.saveTokens(tokens);
      return tokens;
    } on AuthFailure catch (error) {
      if (isAuthoritativeSessionFailure(error)) {
        await _tokenStorage.clearTokens();
      }
      rethrow;
    }
  }

  Future<void> logout() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken != null && refreshToken.isNotEmpty) {
      try {
        await _api.logout(refreshToken: refreshToken);
      } catch (_) {
        // Logout lokal tetap dilakukan agar session perangkat selalu bersih.
      }
    }
    await _tokenStorage.clearTokens();
  }

  UserProfile _validateEmployeeUser(UserProfile user) {
    if (!user.isUser) {
      throw const AuthFailure(
        AuthFailureKind.forbiddenRole,
        'Akun admin tidak dapat masuk aplikasi pegawai.',
      );
    }
    if (user.accountStatus == 'INACTIVE') {
      throw const AuthFailure(
        AuthFailureKind.accountInactive,
        'Akun tidak aktif. Hubungi admin TI.',
      );
    }
    if (user.accountStatus == 'SUSPENDED') {
      throw const AuthFailure(
        AuthFailureKind.accountSuspended,
        'Akun ditangguhkan. Hubungi admin TI.',
      );
    }
    if (!user.isActive) {
      throw const AuthFailure(
        AuthFailureKind.accountInactive,
        'Akun belum dapat digunakan. Hubungi admin TI.',
      );
    }

    return user;
  }
}
