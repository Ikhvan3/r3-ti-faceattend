import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../domain/auth_failure.dart';
import '../domain/auth_token_data.dart';

abstract class TokenStorage {
  Future<void> saveTokens(AuthTokenData tokens);
  Future<String?> readAccessToken();
  Future<String?> readRefreshToken();
  Future<void> clearTokens();
}

abstract class SecureStorageAdapter {
  Future<void> write({required String key, required String? value});
  Future<String?> read({required String key});
  Future<void> delete({required String key});
}

class FlutterSecureStorageAdapter implements SecureStorageAdapter {
  FlutterSecureStorageAdapter([FlutterSecureStorage? storage])
    : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  @override
  Future<void> write({required String key, required String? value}) {
    return _storage.write(key: key, value: value);
  }

  @override
  Future<String?> read({required String key}) {
    return _storage.read(key: key);
  }

  @override
  Future<void> delete({required String key}) {
    return _storage.delete(key: key);
  }
}

class SecureTokenStorage implements TokenStorage {
  SecureTokenStorage({SecureStorageAdapter? storage})
    : _storage = storage ?? FlutterSecureStorageAdapter();

  static const _accessTokenKey = 'r3_access_token';
  static const _refreshTokenKey = 'r3_refresh_token';

  final SecureStorageAdapter _storage;

  @override
  Future<void> saveTokens(AuthTokenData tokens) async {
    try {
      await _storage.write(key: _accessTokenKey, value: tokens.accessToken);
      await _storage.write(key: _refreshTokenKey, value: tokens.refreshToken);
    } catch (_) {
      throw const AuthFailure(
        AuthFailureKind.storageError,
        'Session belum dapat disimpan. Coba lagi.',
      );
    }
  }

  @override
  Future<String?> readAccessToken() async {
    try {
      return _storage.read(key: _accessTokenKey);
    } catch (_) {
      throw const AuthFailure(
        AuthFailureKind.storageError,
        'Session belum dapat dibaca. Silakan login ulang.',
      );
    }
  }

  @override
  Future<String?> readRefreshToken() async {
    try {
      return _storage.read(key: _refreshTokenKey);
    } catch (_) {
      throw const AuthFailure(
        AuthFailureKind.storageError,
        'Session belum dapat dibaca. Silakan login ulang.',
      );
    }
  }

  @override
  Future<void> clearTokens() async {
    try {
      await _storage.delete(key: _accessTokenKey);
      await _storage.delete(key: _refreshTokenKey);
    } catch (_) {
      throw const AuthFailure(
        AuthFailureKind.storageError,
        'Session lokal belum dapat dibersihkan.',
      );
    }
  }
}
