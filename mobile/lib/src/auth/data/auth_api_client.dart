import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;

import '../../config/api_config.dart';
import '../../core/network/api_debug_logger.dart';
import '../domain/auth_failure.dart';
import '../domain/auth_token_data.dart';
import '../domain/user_profile.dart';

abstract class AuthApi {
  Future<AuthTokenData> login({
    required String email,
    required String password,
  });
  Future<AuthTokenData> refresh({required String refreshToken});
  Future<void> logout({required String refreshToken});
  Future<UserProfile> me({required String accessToken});
}

class HttpAuthApiClient implements AuthApi {
  HttpAuthApiClient({
    required String baseUrl,
    required http.Client client,
    Duration timeout = const Duration(seconds: 8),
  }) : _config = ApiConfig(baseUrl: baseUrl),
       _client = client,
       _timeout = timeout;

  final ApiConfig _config;
  final http.Client _client;
  final Duration _timeout;

  @override
  Future<AuthTokenData> login({
    required String email,
    required String password,
  }) async {
    final payload = await _send(
      method: 'POST',
      path: '/auth/login',
      body: <String, Object?>{'email': email, 'password': password},
    );
    return _parseTokenResponse(payload);
  }

  @override
  Future<AuthTokenData> refresh({required String refreshToken}) async {
    final payload = await _send(
      method: 'POST',
      path: '/auth/refresh',
      body: <String, Object?>{'refresh_token': refreshToken},
    );
    return _parseTokenResponse(payload);
  }

  @override
  Future<void> logout({required String refreshToken}) async {
    await _send(
      method: 'POST',
      path: '/auth/logout',
      body: <String, Object?>{'refresh_token': refreshToken},
    );
  }

  @override
  Future<UserProfile> me({required String accessToken}) async {
    final payload = await _send(
      method: 'GET',
      path: '/auth/me',
      accessToken: accessToken,
    );
    final data = payload['data'];
    try {
      return UserProfile.fromJson(data);
    } on FormatException {
      throw const AuthFailure(
        AuthFailureKind.invalidBackendResponse,
        'Respons profil tidak sesuai.',
      );
    }
  }

  Future<Map<String, Object?>> _send({
    required String method,
    required String path,
    Map<String, Object?>? body,
    String? accessToken,
  }) async {
    final headers = <String, String>{
      HttpHeaders.acceptHeader: 'application/json',
      if (body != null) HttpHeaders.contentTypeHeader: 'application/json',
      if (accessToken != null)
        HttpHeaders.authorizationHeader: 'Bearer $accessToken',
    };

    try {
      final request = http.Request(method, Uri.parse(_config.buildUrl(path)));
      request.headers.addAll(headers);
      if (body != null) {
        request.body = jsonEncode(body);
      }

      logApiRequest(method, path);
      final streamed = await _client.send(request).timeout(_timeout);
      final response = await http.Response.fromStream(streamed);
      logApiResponse(method, path, response.statusCode);
      final payload = _decodeObjectResponse(response.body);

      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw _mapHttpError(response.statusCode, payload);
      }
      if (payload['status'] != 'ok') {
        throw const AuthFailure(
          AuthFailureKind.invalidBackendResponse,
          'Respons backend tidak sesuai.',
        );
      }

      return payload;
    } on TimeoutException {
      logApiException(method, path, 'timeout');
      throw const AuthFailure(
        AuthFailureKind.requestTimeout,
        'Request terlalu lama. Coba lagi.',
      );
    } on SocketException {
      logApiException(method, path, 'socket');
      throw const AuthFailure(
        AuthFailureKind.apiUnavailable,
        'Layanan belum tersedia. Coba lagi nanti.',
      );
    } on FormatException {
      logApiException(method, path, 'malformed_response');
      throw const AuthFailure(
        AuthFailureKind.invalidBackendResponse,
        'Respons backend tidak dapat dibaca.',
      );
    } on http.ClientException {
      logApiException(method, path, 'client');
      throw const AuthFailure(
        AuthFailureKind.apiUnavailable,
        'Layanan belum tersedia. Coba lagi nanti.',
      );
    }
  }

  Map<String, Object?> _decodeObjectResponse(String rawBody) {
    final decoded = jsonDecode(rawBody);
    if (decoded is Map) {
      return Map<String, Object?>.from(decoded);
    }
    throw const FormatException('Response bukan object JSON.');
  }

  AuthTokenData _parseTokenResponse(Map<String, Object?> payload) {
    try {
      return AuthTokenData.fromJson(payload['data']);
    } on FormatException {
      throw const AuthFailure(
        AuthFailureKind.invalidBackendResponse,
        'Respons autentikasi tidak sesuai.',
      );
    }
  }

  AuthFailure _mapHttpError(int statusCode, Map<String, Object?> payload) {
    final message = payload['message'] is String
        ? (payload['message'] as String).toLowerCase()
        : '';

    if (statusCode == 401) {
      final isCredentialError =
          message.contains('email') || message.contains('password');
      return AuthFailure(
        isCredentialError
            ? AuthFailureKind.invalidCredentials
            : AuthFailureKind.sessionExpired,
        isCredentialError
            ? 'Email atau password tidak valid.'
            : 'Session berakhir. Silakan login ulang.',
      );
    }
    if (statusCode == 403) {
      return const AuthFailure(
        AuthFailureKind.accountInactive,
        'Akun tidak aktif. Hubungi admin TI.',
      );
    }
    if (statusCode >= 500) {
      return const AuthFailure(
        AuthFailureKind.internalError,
        'Layanan mengalami gangguan. Coba lagi nanti.',
      );
    }

    return const AuthFailure(
      AuthFailureKind.invalidBackendResponse,
      'Request tidak dapat diproses.',
    );
  }
}
