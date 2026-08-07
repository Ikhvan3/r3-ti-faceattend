import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;

import '../../auth/data/auth_api_client.dart';
import '../../auth/data/token_storage.dart';
import '../../auth/domain/auth_failure.dart';
import '../../config/api_config.dart';
import 'api_debug_logger.dart';

enum AuthenticatedApiFailureKind {
  sessionExpired,
  apiUnavailable,
  requestTimeout,
  malformedResponse,
}

class AuthenticatedApiFailure implements Exception {
  const AuthenticatedApiFailure(this.kind, this.message);

  final AuthenticatedApiFailureKind kind;
  final String message;

  @override
  String toString() => message;
}

class AuthenticatedApiResponse {
  const AuthenticatedApiResponse({
    required this.statusCode,
    required this.payload,
  });

  final int statusCode;
  final Map<String, Object?> payload;
}

abstract class AuthenticatedRequester {
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
    Map<String, Object?>? body,
  });
}

class AuthenticatedApiClient implements AuthenticatedRequester {
  AuthenticatedApiClient({
    required String baseUrl,
    required http.Client client,
    required TokenStorage tokenStorage,
    required AuthApi authApi,
    Duration timeout = const Duration(seconds: 8),
  }) : _config = ApiConfig(baseUrl: baseUrl),
       _client = client,
       _tokenStorage = tokenStorage,
       _authApi = authApi,
       _timeout = timeout;

  final ApiConfig _config;
  final http.Client _client;
  final TokenStorage _tokenStorage;
  final AuthApi _authApi;
  final Duration _timeout;
  Future<String>? _refreshInFlight;

  @override
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
    Map<String, Object?>? body,
  }) async {
    final accessToken = await _tokenStorage.readAccessToken();
    if (accessToken == null || accessToken.isEmpty) {
      await _tokenStorage.clearTokens();
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    }

    final first = await _sendOnce(
      method: method,
      path: path,
      accessToken: accessToken,
      queryParameters: queryParameters,
      body: body,
    );
    if (first.statusCode != HttpStatus.unauthorized) {
      return first;
    }

    final rotatedAccessToken = await _refreshSingleFlight(accessToken);
    return _sendOnce(
      method: method,
      path: path,
      accessToken: rotatedAccessToken,
      queryParameters: queryParameters,
      body: body,
    );
  }

  Future<String> _refreshSingleFlight(String tokenUsedByRequest) {
    final inFlight = _refreshInFlight;
    if (inFlight != null) {
      return inFlight;
    }

    final future = _refreshIfStillCurrent(tokenUsedByRequest);
    _refreshInFlight = future;
    unawaited(
      future.then<void>((_) {}, onError: (_) {}).whenComplete(() {
        if (identical(_refreshInFlight, future)) {
          _refreshInFlight = null;
        }
      }),
    );
    return future;
  }

  Future<String> _refreshIfStillCurrent(String tokenUsedByRequest) async {
    final currentAccessToken = await _tokenStorage.readAccessToken();
    if (currentAccessToken != null &&
        currentAccessToken.isNotEmpty &&
        currentAccessToken != tokenUsedByRequest) {
      return currentAccessToken;
    }

    return _refreshOnce();
  }

  Future<AuthenticatedApiResponse> _sendOnce({
    required String method,
    required String path,
    required String accessToken,
    Map<String, String>? queryParameters,
    Map<String, Object?>? body,
  }) async {
    try {
      final uri = _uri(path, queryParameters);
      final request = http.Request(method, uri);
      request.headers.addAll(<String, String>{
        HttpHeaders.acceptHeader: 'application/json',
        HttpHeaders.authorizationHeader: 'Bearer $accessToken',
      });
      if (body != null) {
        request.headers[HttpHeaders.contentTypeHeader] = 'application/json';
        request.body = jsonEncode(body);
      }

      logApiRequest(method, path);
      final streamed = await _client.send(request).timeout(_timeout);
      final response = await http.Response.fromStream(streamed);
      logApiResponse(method, path, response.statusCode);
      return AuthenticatedApiResponse(
        statusCode: response.statusCode,
        payload: _decodeObjectResponse(response.body),
      );
    } on TimeoutException {
      logApiException(method, path, 'timeout');
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.requestTimeout,
        'Request terlalu lama. Coba lagi.',
      );
    } on SocketException {
      logApiException(method, path, 'socket');
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.apiUnavailable,
        'Layanan belum tersedia. Coba lagi nanti.',
      );
    } on FormatException {
      logApiException(method, path, 'malformed_response');
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.malformedResponse,
        'Respons backend tidak dapat dibaca.',
      );
    } on http.ClientException {
      logApiException(method, path, 'client');
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.apiUnavailable,
        'Layanan belum tersedia. Coba lagi nanti.',
      );
    }
  }

  Future<String> _refreshOnce() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      await _tokenStorage.clearTokens();
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    }

    try {
      final tokens = await _authApi.refresh(refreshToken: refreshToken);
      await _tokenStorage.saveTokens(tokens);
      return tokens.accessToken;
    } on AuthFailure catch (error) {
      if (isAuthoritativeSessionFailure(error)) {
        await _tokenStorage.clearTokens();
        throw const AuthenticatedApiFailure(
          AuthenticatedApiFailureKind.sessionExpired,
          'Session berakhir. Silakan login kembali.',
        );
      }
      if (error.kind == AuthFailureKind.requestTimeout) {
        throw const AuthenticatedApiFailure(
          AuthenticatedApiFailureKind.requestTimeout,
          'Koneksi terlalu lambat. Silakan coba lagi.',
        );
      }
      if (error.kind == AuthFailureKind.apiUnavailable ||
          error.kind == AuthFailureKind.internalError) {
        throw const AuthenticatedApiFailure(
          AuthenticatedApiFailureKind.apiUnavailable,
          'Layanan belum tersedia. Pastikan backend aktif dan perangkat terhubung.',
        );
      }
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.malformedResponse,
        'Respons backend tidak dapat dibaca.',
      );
    }
  }

  Uri _uri(String path, Map<String, String>? queryParameters) {
    final built = Uri.parse(_config.buildUrl(path));
    if (queryParameters == null || queryParameters.isEmpty) {
      return built;
    }
    return built.replace(queryParameters: queryParameters);
  }

  Map<String, Object?> _decodeObjectResponse(String rawBody) {
    final decoded = jsonDecode(rawBody);
    if (decoded is Map) {
      return Map<String, Object?>.from(decoded);
    }
    throw const FormatException('Response bukan object JSON.');
  }
}
