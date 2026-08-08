import 'dart:io';

import '../../core/network/authenticated_api_client.dart';
import '../domain/face_attendance_grant.dart';
import '../domain/face_failure.dart';
import '../domain/face_status.dart';
import '../domain/face_verification_result.dart';

abstract class FaceApi {
  Future<FaceStatus> getStatus();
  Future<FaceStatus> enroll({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  });
  Future<FaceVerificationResult> verify({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  });
  Future<FaceAttendanceGrant> verifyForAttendance({
    required String purpose,
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  });
}

class HttpFaceApiClient implements FaceApi {
  const HttpFaceApiClient({required AuthenticatedRequester client})
    : _client = client;

  final AuthenticatedRequester _client;

  @override
  Future<FaceStatus> getStatus() async {
    final response = await _send(method: 'GET', path: '/face/status');
    return _parseStatus(response);
  }

  @override
  Future<FaceStatus> enroll({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) async {
    final response = await _send(
      method: 'POST',
      path: '/face/enroll',
      body: <String, Object?>{
        'embedding': embedding,
        'embedding_model': embeddingModel,
        'embedding_version': embeddingVersion,
      },
    );
    return _parseStatus(response);
  }

  @override
  Future<FaceVerificationResult> verify({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) async {
    final response = await _send(
      method: 'POST',
      path: '/face/verify',
      body: <String, Object?>{
        'embedding': embedding,
        'embedding_model': embeddingModel,
        'embedding_version': embeddingVersion,
      },
    );
    try {
      return FaceVerificationResult.fromJson(response.payload['data']);
    } on FormatException {
      throw const FaceFailure(
        FaceFailureKind.malformedResponse,
        'Respons verifikasi wajah tidak sesuai.',
      );
    }
  }

  @override
  Future<FaceAttendanceGrant> verifyForAttendance({
    required String purpose,
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) async {
    final response = await _send(
      method: 'POST',
      path: '/face/verify-for-attendance',
      body: <String, Object?>{
        'purpose': purpose,
        'embedding': embedding,
        'embedding_model': embeddingModel,
        'embedding_version': embeddingVersion,
      },
    );
    try {
      return FaceAttendanceGrant.fromJson(response.payload['data']);
    } on FormatException {
      throw const FaceFailure(
        FaceFailureKind.malformedResponse,
        'Respons verifikasi wajah tidak sesuai.',
      );
    }
  }

  Future<AuthenticatedApiResponse> _send({
    required String method,
    required String path,
    Map<String, Object?>? body,
  }) async {
    try {
      final response = await _client.send(
        method: method,
        path: path,
        body: body,
      );
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw _mapStatus(response.statusCode, path, response.payload);
      }
      if (response.payload['status'] != 'ok') {
        throw const FaceFailure(
          FaceFailureKind.malformedResponse,
          'Respons face enrollment tidak sesuai.',
        );
      }
      return response;
    } on AuthenticatedApiFailure catch (error) {
      throw _mapAuthenticatedFailure(error);
    }
  }

  FaceStatus _parseStatus(AuthenticatedApiResponse response) {
    try {
      return FaceStatus.fromJson(response.payload['data']);
    } on FormatException {
      throw const FaceFailure(
        FaceFailureKind.malformedResponse,
        'Respons status wajah tidak sesuai.',
      );
    }
  }

  FaceFailure _mapStatus(
    int statusCode,
    String path,
    Map<String, Object?> payload,
  ) {
    final message = payload['message'] is String
        ? (payload['message'] as String).toLowerCase()
        : '';

    if (statusCode == HttpStatus.unauthorized) {
      return const FaceFailure(
        FaceFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    }
    if (statusCode == HttpStatus.forbidden) {
      if (path == '/face/verify-for-attendance') {
        return const FaceFailure(
          FaceFailureKind.verificationRejected,
          'Wajah tidak sesuai dengan data yang terdaftar.',
        );
      }
      return const FaceFailure(
        FaceFailureKind.accountForbidden,
        'Akun tidak diizinkan melakukan enrollment wajah.',
      );
    }
    if (statusCode == HttpStatus.conflict) {
      if (path == '/face/verify') {
        return const FaceFailure(
          FaceFailureKind.notEnrolled,
          'Wajah belum terdaftar.',
        );
      }
      if (path == '/face/enroll' && message.contains('akun lain')) {
        return const FaceFailure(
          FaceFailureKind.duplicateEnrollment,
          'Wajah ini telah terdaftar pada akun lain. Hubungi administrator jika data ini keliru.',
        );
      }
      return const FaceFailure(
        FaceFailureKind.duplicateEnrollment,
        'Wajah sudah terdaftar. Hubungi administrator jika enrollment perlu diubah.',
      );
    }
    if (statusCode == HttpStatus.badRequest) {
      return const FaceFailure(
        FaceFailureKind.invalidEmbedding,
        'Data wajah tidak valid.',
      );
    }
    if (statusCode >= 500) {
      return const FaceFailure(
        FaceFailureKind.internalError,
        'Layanan face enrollment mengalami gangguan.',
      );
    }
    return const FaceFailure(
      FaceFailureKind.malformedResponse,
      'Request face enrollment tidak dapat diproses.',
    );
  }

  FaceFailure _mapAuthenticatedFailure(AuthenticatedApiFailure error) {
    switch (error.kind) {
      case AuthenticatedApiFailureKind.sessionExpired:
        return FaceFailure(FaceFailureKind.sessionExpired, error.message);
      case AuthenticatedApiFailureKind.apiUnavailable:
        return FaceFailure(FaceFailureKind.apiUnavailable, error.message);
      case AuthenticatedApiFailureKind.requestTimeout:
        return FaceFailure(FaceFailureKind.requestTimeout, error.message);
      case AuthenticatedApiFailureKind.malformedResponse:
        return FaceFailure(FaceFailureKind.malformedResponse, error.message);
    }
  }
}
