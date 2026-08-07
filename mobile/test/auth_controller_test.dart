import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/auth/data/auth_repository.dart';
import 'package:r3_ti_faceattend/src/auth/presentation/auth_controller.dart';

import 'auth_test_fakes.dart';

void main() {
  test('initial ke authenticated', () async {
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final controller = AuthController(
      AuthRepository(api: FakeAuthApi(), tokenStorage: storage),
    );

    await controller.initialize();

    expect(controller.status, AuthControllerStatus.authenticated);
    expect(controller.user?.role, 'USER');
  });

  test('initial ke unauthenticated', () async {
    final controller = AuthController(
      AuthRepository(api: FakeAuthApi(), tokenStorage: FakeTokenStorage()),
    );

    await controller.initialize();

    expect(controller.status, AuthControllerStatus.unauthenticated);
    expect(controller.sessionRestoreFailed, isFalse);
  });

  test('initial gagal sementara bisa dicoba ulang', () async {
    final api = FakeAuthApi()..meError = apiUnavailableFailure;
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final controller = AuthController(
      AuthRepository(api: api, tokenStorage: storage),
    );

    await controller.initialize();

    expect(controller.status, AuthControllerStatus.failure);
    expect(controller.sessionRestoreFailed, isTrue);
    expect(storage.accessToken, 'access-token');

    api.meError = null;
    await controller.initialize();

    expect(controller.status, AuthControllerStatus.authenticated);
    expect(controller.sessionRestoreFailed, isFalse);
  });

  test('login loading ke authenticated', () async {
    final controller = AuthController(
      AuthRepository(api: FakeAuthApi(), tokenStorage: FakeTokenStorage()),
    );

    final future = controller.login(
      email: 'pegawai@example.test',
      password: 'password',
    );
    expect(controller.status, AuthControllerStatus.loading);
    await future;

    expect(controller.status, AuthControllerStatus.authenticated);
    expect(controller.user?.name, 'Pegawai Dummy');
  });

  test('login gagal menjadi failure', () async {
    final api = FakeAuthApi()..loginError = invalidCredentialFailure;
    final controller = AuthController(
      AuthRepository(api: api, tokenStorage: FakeTokenStorage()),
    );

    await controller.login(email: 'pegawai@example.test', password: 'wrong');

    expect(controller.status, AuthControllerStatus.failure);
    expect(controller.errorMessage, isNotNull);
    expect(controller.sessionRestoreFailed, isFalse);
  });

  test('multiple submit dicegah', () async {
    final api = FakeAuthApi()..loginCompleter = Completer<void>();
    final controller = AuthController(
      AuthRepository(api: api, tokenStorage: FakeTokenStorage()),
    );

    final first = controller.login(
      email: 'a@example.test',
      password: 'password',
    );
    final second = controller.login(
      email: 'b@example.test',
      password: 'password',
    );
    api.loginCompleter?.complete();
    await Future.wait(<Future<void>>[first, second]);

    expect(api.loginCalls, 1);
  });

  test('logout menjadi unauthenticated', () async {
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';
    final controller = AuthController(
      AuthRepository(api: FakeAuthApi(), tokenStorage: storage),
    );

    await controller.initialize();
    await controller.logout();

    expect(controller.status, AuthControllerStatus.unauthenticated);
    expect(controller.user, isNull);
  });
}
