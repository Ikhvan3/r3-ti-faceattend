import 'dart:async';

import 'package:r3_ti_faceattend/src/auth/data/auth_api_client.dart';
import 'package:r3_ti_faceattend/src/auth/data/token_storage.dart';
import 'package:r3_ti_faceattend/src/auth/domain/auth_failure.dart';
import 'package:r3_ti_faceattend/src/auth/domain/auth_token_data.dart';
import 'package:r3_ti_faceattend/src/auth/domain/user_profile.dart';

class FakeAuthApi implements AuthApi {
  AuthTokenData? loginResult;
  AuthTokenData? refreshResult;
  UserProfile? meResult;
  Object? loginError;
  Object? refreshError;
  Object? meError;
  Object? logoutError;
  Completer<void>? loginCompleter;
  final List<Object> meErrors = <Object>[];
  final List<UserProfile> meResults = <UserProfile>[];
  int loginCalls = 0;
  int refreshCalls = 0;
  int meCalls = 0;
  int logoutCalls = 0;

  @override
  Future<AuthTokenData> login({
    required String email,
    required String password,
  }) async {
    loginCalls++;
    await loginCompleter?.future;
    if (loginError != null) {
      throw loginError!;
    }
    return loginResult ?? userTokens();
  }

  @override
  Future<AuthTokenData> refresh({required String refreshToken}) async {
    refreshCalls++;
    if (refreshError != null) {
      throw refreshError!;
    }
    return refreshResult ??
        userTokens(access: 'new-access', refresh: 'new-refresh');
  }

  @override
  Future<void> logout({required String refreshToken}) async {
    logoutCalls++;
    if (logoutError != null) {
      throw logoutError!;
    }
  }

  @override
  Future<UserProfile> me({required String accessToken}) async {
    meCalls++;
    if (meErrors.isNotEmpty) {
      throw meErrors.removeAt(0);
    }
    if (meResults.isNotEmpty) {
      return meResults.removeAt(0);
    }
    if (meError != null) {
      throw meError!;
    }
    return meResult ?? userProfile();
  }
}

class FakeTokenStorage implements TokenStorage {
  String? accessToken;
  String? refreshToken;
  int saveCalls = 0;
  int clearCalls = 0;

  @override
  Future<void> saveTokens(AuthTokenData tokens) async {
    saveCalls++;
    accessToken = tokens.accessToken;
    refreshToken = tokens.refreshToken;
  }

  @override
  Future<String?> readAccessToken() async => accessToken;

  @override
  Future<String?> readRefreshToken() async => refreshToken;

  @override
  Future<void> clearTokens() async {
    clearCalls++;
    accessToken = null;
    refreshToken = null;
  }
}

class FakeSecureStorageAdapter implements SecureStorageAdapter {
  final Map<String, String> values = <String, String>{};
  bool throwOnWrite = false;
  bool throwOnRead = false;
  bool throwOnDelete = false;

  @override
  Future<void> write({required String key, required String? value}) async {
    if (throwOnWrite) {
      throw StateError('write failed');
    }
    if (value == null) {
      values.remove(key);
      return;
    }
    values[key] = value;
  }

  @override
  Future<String?> read({required String key}) async {
    if (throwOnRead) {
      throw StateError('read failed');
    }
    return values[key];
  }

  @override
  Future<void> delete({required String key}) async {
    if (throwOnDelete) {
      throw StateError('delete failed');
    }
    values.remove(key);
  }
}

UserProfile userProfile({
  String role = 'USER',
  String status = 'ACTIVE',
  String name = 'Pegawai Dummy',
}) {
  return UserProfile(
    id: '00000000-0000-4000-8000-000000000001',
    employeeNumber: 'EMP-001',
    name: name,
    email: 'pegawai.dummy@example.test',
    phone: '081234567890',
    position: 'Staf TI',
    role: role,
    accountStatus: status,
  );
}

AuthTokenData userTokens({
  String access = 'access-token',
  String refresh = 'refresh-token',
  UserProfile? user,
}) {
  return AuthTokenData(
    accessToken: access,
    refreshToken: refresh,
    tokenType: 'Bearer',
    expiresIn: 900,
    user: user ?? userProfile(),
  );
}

const expiredFailure = AuthFailure(
  AuthFailureKind.sessionExpired,
  'Session berakhir.',
);

const invalidCredentialFailure = AuthFailure(
  AuthFailureKind.invalidCredentials,
  'Email atau password tidak valid.',
);
