import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;

import '../../auth/data/auth_api_client.dart';
import '../../auth/data/token_storage.dart';
import '../../auth/domain/auth_failure.dart';
import '../../config/api_config.dart';

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

  @override
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
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
    );
    if (first.statusCode != HttpStatus.unauthorized) {
      return first;
    }

    final rotatedAccessToken = await _refreshOnce();
    return _sendOnce(
      method: method,
      path: path,
      accessToken: rotatedAccessToken,
      queryParameters: queryParameters,
    );
  }

  Future<AuthenticatedApiResponse> _sendOnce({
    required String method,
    required String path,
    required String accessToken,
    Map<String, String>? queryParameters,
  }) async {
    try {
      final uri = _uri(path, queryParameters);
      final request = http.Request(method, uri);
      request.headers.addAll(<String, String>{
        HttpHeaders.acceptHeader: 'application/json',
        HttpHeaders.authorizationHeader: 'Bearer $accessToken',
      });

      final streamed = await _client.send(request).timeout(_timeout);
      final response = await http.Response.fromStream(streamed);
      return AuthenticatedApiResponse(
        statusCode: response.statusCode,
        payload: _decodeObjectResponse(response.body),
      );
    } on TimeoutException {
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.requestTimeout,
        'Request terlalu lama. Coba lagi.',
      );
    } on SocketException {
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.apiUnavailable,
        'Layanan belum tersedia. Coba lagi nanti.',
      );
    } on FormatException {
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.malformedResponse,
        'Respons backend tidak dapat dibaca.',
      );
    } on http.ClientException {
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
    } on AuthFailure {
      await _tokenStorage.clearTokens();
      throw const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
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
