import 'dart:async';
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

  test('refresh gagal sementara mempertahankan session', () async {
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
      authApi: FakeAuthApi()..refreshError = apiUnavailableFailure,
    );

    await expectLater(
      client.send(method: 'GET', path: '/attendance/today'),
      throwsA(isA<AuthenticatedApiFailure>()),
    );
    expect(storage.accessToken, 'old-access');
    expect(storage.refreshToken, 'old-refresh');
  });

  test('401 paralel berbagi satu refresh token', () async {
    var requestCalls = 0;
    final storage = FakeTokenStorage()
      ..accessToken = 'old-access'
      ..refreshToken = 'old-refresh';
    final authApi = FakeAuthApi()
      ..refreshCompleter = Completer<void>()
      ..refreshResult = userTokens(
        access: 'new-access',
        refresh: 'new-refresh',
      );
    final httpClient = MockClient((request) async {
      requestCalls++;
      if (request.headers['authorization'] == 'Bearer old-access') {
        return http.Response(
          jsonEncode(<String, Object?>{'status': 'error'}),
          401,
        );
      }
      return http.Response(jsonEncode(<String, Object?>{'status': 'ok'}), 200);
    });
    final client = AuthenticatedApiClient(
      baseUrl: 'http://127.0.0.1:8080/api/v1',
      client: httpClient,
      tokenStorage: storage,
      authApi: authApi,
    );

    final first = client.send(method: 'GET', path: '/attendance/today');
    final second = client.send(method: 'GET', path: '/attendance/history');
    await Future<void>.delayed(Duration.zero);
    authApi.refreshCompleter?.complete();
    final responses = await Future.wait(<Future<AuthenticatedApiResponse>>[
      first,
      second,
    ]);

    expect(responses.map((response) => response.statusCode), everyElement(200));
    expect(requestCalls, 4);
    expect(authApi.refreshCalls, 1);
    expect(storage.accessToken, 'new-access');
    expect(storage.refreshToken, 'new-refresh');
  });

  test(
    '401 memakai token terbaru jika refresh sudah dilakukan request lain',
    () async {
      var requestCalls = 0;
      final storage = FakeTokenStorage()
        ..accessToken = 'old-access'
        ..refreshToken = 'old-refresh';
      final httpClient = MockClient((request) async {
        requestCalls++;
        if (request.headers['authorization'] == 'Bearer old-access') {
          storage.accessToken = 'new-access';
          return http.Response(
            jsonEncode(<String, Object?>{'status': 'error'}),
            401,
          );
        }
        expect(request.headers['authorization'], 'Bearer new-access');
        return http.Response(
          jsonEncode(<String, Object?>{'status': 'ok'}),
          200,
        );
      });
      final authApi = FakeAuthApi();
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
      expect(requestCalls, 2);
      expect(authApi.refreshCalls, 0);
      expect(storage.accessToken, 'new-access');
      expect(storage.refreshToken, 'old-refresh');
    },
  );

  test('mengirim JSON body saat tersedia', () async {
    late http.Request capturedRequest;
    final httpClient = MockClient((request) async {
      capturedRequest = request;
      return http.Response(jsonEncode(<String, Object?>{'status': 'ok'}), 200);
    });
    final storage = FakeTokenStorage()..accessToken = 'access-token';
    final client = AuthenticatedApiClient(
      baseUrl: 'http://127.0.0.1:8080/api/v1',
      client: httpClient,
      tokenStorage: storage,
      authApi: FakeAuthApi(),
    );

    await client.send(
      method: 'POST',
      path: '/attendance/check-in',
      body: <String, Object?>{
        'latitude': -6.98946,
        'longitude': 110.416735,
        'accuracy_meters': 12.5,
      },
    );

    expect(capturedRequest.headers['content-type'], 'application/json');
    expect(jsonDecode(capturedRequest.body), <String, Object?>{
      'latitude': -6.98946,
      'longitude': 110.416735,
      'accuracy_meters': 12.5,
    });
  });
}
