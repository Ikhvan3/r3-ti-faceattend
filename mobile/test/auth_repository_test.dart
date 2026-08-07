import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/auth/data/auth_repository.dart';
import 'package:r3_ti_faceattend/src/auth/domain/auth_failure.dart';

import 'auth_test_fakes.dart';

void main() {
  test('login USER berhasil dan menyimpan token', () async {
    final api = FakeAuthApi();
    final storage = FakeTokenStorage();
    final repository = AuthRepository(api: api, tokenStorage: storage);

    final user = await repository.login(
      email: 'pegawai@example.test',
      password: 'password',
    );

    expect(user.role, 'USER');
    expect(storage.accessToken, 'access-token');
    expect(storage.refreshToken, 'refresh-token');
    expect(storage.saveCalls, 1);
    expect(api.loginCalls, 1);
    expect(api.meCalls, 1);
  });

  test('login ADMIN ditolak dan token tidak disimpan', () async {
    final api = FakeAuthApi()..meResult = userProfile(role: 'ADMIN');
    final storage = FakeTokenStorage();
    final repository = AuthRepository(api: api, tokenStorage: storage);

    expect(
      () => repository.login(email: 'admin@example.test', password: 'password'),
      throwsA(isA<AuthFailure>()),
    );
    expect(storage.saveCalls, 0);
  });

  test('akun INACTIVE ditolak', () async {
    final api = FakeAuthApi()..meResult = userProfile(status: 'INACTIVE');
    final repository = AuthRepository(
      api: api,
      tokenStorage: FakeTokenStorage(),
    );

    expect(
      () => repository.login(email: 'u@example.test', password: 'password'),
      throwsA(isA<AuthFailure>()),
    );
  });

  test('akun SUSPENDED ditolak', () async {
    final api = FakeAuthApi()..meResult = userProfile(status: 'SUSPENDED');
    final repository = AuthRepository(
      api: api,
      tokenStorage: FakeTokenStorage(),
    );

    expect(
      () => repository.login(email: 'u@example.test', password: 'password'),
      throwsA(isA<AuthFailure>()),
    );
  });

  test('token tidak disimpan jika login gagal', () async {
    final api = FakeAuthApi()..loginError = invalidCredentialFailure;
    final storage = FakeTokenStorage();
    final repository = AuthRepository(api: api, tokenStorage: storage);

    expect(
      () => repository.login(email: 'u@example.test', password: 'wrong'),
      throwsA(isA<AuthFailure>()),
    );
    expect(storage.saveCalls, 0);
  });

  test('token tidak disimpan jika profil me gagal setelah login', () async {
    final api = FakeAuthApi()..meError = expiredFailure;
    final storage = FakeTokenStorage();
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await expectLater(
      repository.login(email: 'u@example.test', password: 'password'),
      throwsA(isA<AuthFailure>()),
    );

    expect(api.loginCalls, 1);
    expect(api.meCalls, 1);
    expect(storage.saveCalls, 0);
    expect(storage.accessToken, isNull);
  });

  test('restore access token berhasil', () async {
    final api = FakeAuthApi();
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    final user = await repository.restoreSession();

    expect(user?.name, 'Pegawai Dummy');
    expect(api.meCalls, 1);
    expect(api.refreshCalls, 0);
  });

  test('access token expired melakukan refresh satu kali', () async {
    final api = FakeAuthApi()
      ..meErrors.add(expiredFailure)
      ..refreshResult = userTokens(
        access: 'rotated-access',
        refresh: 'rotated-refresh',
      );
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    final user = await repository.restoreSession();

    expect(user?.role, 'USER');
    expect(api.refreshCalls, 1);
    expect(api.meCalls, 2);
    expect(storage.accessToken, 'rotated-access');
    expect(storage.refreshToken, 'rotated-refresh');
  });

  test('refresh token lama diganti', () async {
    final api = FakeAuthApi()
      ..refreshResult = userTokens(
        access: 'new-access',
        refresh: 'new-refresh',
      );
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await repository.refreshSession();

    expect(storage.accessToken, 'new-access');
    expect(storage.refreshToken, 'new-refresh');
  });

  test('refresh gagal membersihkan token', () async {
    final api = FakeAuthApi()..refreshError = expiredFailure;
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await expectLater(repository.refreshSession(), throwsA(isA<AuthFailure>()));
    expect(storage.accessToken, isNull);
    expect(storage.refreshToken, isNull);
  });

  test('refresh gagal sementara tidak membersihkan token', () async {
    final api = FakeAuthApi()..refreshError = apiUnavailableFailure;
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await expectLater(repository.refreshSession(), throwsA(isA<AuthFailure>()));
    expect(storage.accessToken, 'old-access');
    expect(storage.refreshToken, 'old-refresh');
  });

  test('restore gagal sementara mempertahankan token', () async {
    final api = FakeAuthApi()..meError = apiUnavailableFailure;
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await expectLater(repository.restoreSession(), throwsA(isA<AuthFailure>()));

    expect(storage.accessToken, 'access-token');
    expect(storage.refreshToken, 'refresh-token');
  });

  test('login akun berbeda mengganti token lama', () async {
    final api = FakeAuthApi()
      ..loginResult = userTokens(access: 'access-b', refresh: 'refresh-b')
      ..meResult = userProfile(name: 'Pegawai B');
    final storage = FakeTokenStorage()
      ..accessToken = 'access-a'
      ..refreshToken = 'refresh-a';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await repository.logout();
    final user = await repository.login(
      email: 'pegawai.b@example.test',
      password: 'password',
    );

    expect(user.name, 'Pegawai B');
    expect(storage.accessToken, 'access-b');
    expect(storage.refreshToken, 'refresh-b');
  });

  test('logout membersihkan token walaupun API gagal', () async {
    final api = FakeAuthApi()..logoutError = Exception('network');
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final repository = AuthRepository(api: api, tokenStorage: storage);

    await repository.logout();

    expect(api.logoutCalls, 1);
    expect(storage.accessToken, isNull);
    expect(storage.refreshToken, isNull);
  });
}
