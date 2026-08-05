import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:r3_ti_faceattend/src/core/network/authenticated_api_client.dart';

import 'auth_test_fakes.dart';

void main() {
  test('401 melakukan refresh satu kali dan menyimpan token baru', () async {
    var calls = 0;
    final httpClient = MockClient((request) async {
      calls++;
      if (calls == 1) {
        expect(request.headers['authorization'], 'Bearer old-access');
        return http.Response(
          jsonEncode(<String, Object?>{'status': 'error'}),
          401,
        );
      }
      expect(request.headers['authorization'], 'Bearer new-access');
      return http.Response(
        jsonEncode(<String, Object?>{
          'status': 'ok',
          'data': <String, Object?>{},
        }),
        200,
      );
    });
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final authApi = FakeAuthApi()
      ..refreshResult = userTokens(
        access: 'new-access',
        refresh: 'new-refresh',
      );
    final client = AuthenticatedApiClient(
      baseUrl: 'http://127.0.0.1:8080/api/v1',
      client: httpClient,
      tokenStorage: storage,
      authApi: authApi,
    );

    final response = await client.send(
      method: 'GET',
      path: '/attendance/today',
    );

    expect(response.statusCode, 200);
    expect(calls, 2);
    expect(authApi.refreshCalls, 1);
    expect(storage.accessToken, 'new-access');
    expect(storage.refreshToken, 'new-refresh');
  });

  test('refresh gagal membersihkan session', () async {
    final httpClient = MockClient((_) async {
      return http.Response(
        jsonEncode(<String, Object?>{'status': 'error'}),
        401,
      );
    });
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final client = AuthenticatedApiClient(
      baseUrl: 'http://127.0.0.1:8080/api/v1',
      client: httpClient,
      tokenStorage: storage,
      authApi: FakeAuthApi()..refreshError = expiredFailure,
    );

    await expectLater(
      client.send(method: 'GET', path: '/attendance/today'),
      throwsA(isA<AuthenticatedApiFailure>()),
    );
    expect(storage.accessToken, isNull);
    expect(storage.refreshToken, isNull);
  });
}
