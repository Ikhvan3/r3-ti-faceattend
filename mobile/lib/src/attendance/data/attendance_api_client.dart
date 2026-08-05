import 'dart:io';

import '../../core/network/authenticated_api_client.dart';
import '../domain/attendance_failure.dart';
import '../domain/attendance_models.dart';

abstract class AttendanceApi {
  Future<AttendanceToday> getTodayAttendance();
  Future<AttendanceToday> checkIn();
  Future<AttendanceToday> checkOut();
  Future<AttendanceHistoryResponse> getAttendanceHistory({
    required int page,
    required int pageSize,
  });
}

class HttpAttendanceApiClient implements AttendanceApi {
  const HttpAttendanceApiClient({required AuthenticatedRequester client})
    : _client = client;

  final AuthenticatedRequester _client;

  @override
  Future<AttendanceToday> getTodayAttendance() async {
    final response = await _send(method: 'GET', path: '/attendance/today');
    return _parseToday(response);
  }

  @override
  Future<AttendanceToday> checkIn() async {
    final response = await _send(method: 'POST', path: '/attendance/check-in');
    return _parseToday(response);
  }

  @override
  Future<AttendanceToday> checkOut() async {
    final response = await _send(method: 'POST', path: '/attendance/check-out');
    return _parseToday(response);
  }

  @override
  Future<AttendanceHistoryResponse> getAttendanceHistory({
    required int page,
    required int pageSize,
  }) async {
    final response = await _send(
      method: 'GET',
      path: '/attendance/history',
      queryParameters: <String, String>{
        'page': page.toString(),
        'page_size': pageSize.toString(),
      },
    );
    try {
      return AttendanceHistoryResponse.fromJson(response.payload['data']);
    } on FormatException {
      throw const AttendanceFailure(
        AttendanceFailureKind.malformedResponse,
        'Respons riwayat absensi tidak sesuai.',
      );
    }
  }

  Future<AuthenticatedApiResponse> _send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
  }) async {
    try {
      final response = await _client.send(
        method: method,
        path: path,
        queryParameters: queryParameters,
      );
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw _mapStatus(response.statusCode, response.payload);
      }
      if (response.payload['status'] != 'ok') {
        throw const AttendanceFailure(
          AttendanceFailureKind.malformedResponse,
          'Respons backend tidak sesuai.',
        );
      }
      return response;
    } on AuthenticatedApiFailure catch (error) {
      throw _mapAuthenticatedFailure(error);
    }
  }

  AttendanceToday _parseToday(AuthenticatedApiResponse response) {
    try {
      return AttendanceToday.fromJson(response.payload['data']);
    } on FormatException {
      throw const AttendanceFailure(
        AttendanceFailureKind.malformedResponse,
        'Respons absensi hari ini tidak sesuai.',
      );
    }
  }

  AttendanceFailure _mapStatus(int statusCode, Map<String, Object?> payload) {
    final message = payload['message'] is String
        ? (payload['message'] as String).toLowerCase()
        : '';

    if (statusCode == HttpStatus.unauthorized) {
      return const AttendanceFailure(
        AttendanceFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    }
    if (statusCode == HttpStatus.forbidden) {
      return const AttendanceFailure(
        AttendanceFailureKind.accountForbidden,
        'Akun tidak diizinkan melakukan absensi.',
      );
    }
    if (statusCode == HttpStatus.notFound) {
      return const AttendanceFailure(
        AttendanceFailureKind.scheduleUnavailable,
        'Jadwal kerja belum tersedia. Hubungi administrator.',
      );
    }
    if (statusCode == HttpStatus.conflict) {
      if (message.contains('belum check-in')) {
        return const AttendanceFailure(
          AttendanceFailureKind.notCheckedIn,
          'Anda harus check-in sebelum melakukan check-out.',
        );
      }
      if (message.contains('check-out')) {
        return const AttendanceFailure(
          AttendanceFailureKind.alreadyCheckedOut,
          'Anda sudah melakukan check-out hari ini.',
        );
      }
      return const AttendanceFailure(
        AttendanceFailureKind.alreadyCheckedIn,
        'Anda sudah melakukan check-in hari ini.',
      );
    }
    if (statusCode >= 500) {
      return const AttendanceFailure(
        AttendanceFailureKind.internalError,
        'Layanan absensi mengalami gangguan. Coba lagi nanti.',
      );
    }

    return const AttendanceFailure(
      AttendanceFailureKind.malformedResponse,
      'Request absensi tidak dapat diproses.',
    );
  }

  AttendanceFailure _mapAuthenticatedFailure(AuthenticatedApiFailure error) {
    switch (error.kind) {
      case AuthenticatedApiFailureKind.sessionExpired:
        return AttendanceFailure(
          AttendanceFailureKind.sessionExpired,
          error.message,
        );
      case AuthenticatedApiFailureKind.apiUnavailable:
        return AttendanceFailure(
          AttendanceFailureKind.apiUnavailable,
          error.message,
        );
      case AuthenticatedApiFailureKind.requestTimeout:
        return AttendanceFailure(
          AttendanceFailureKind.requestTimeout,
          error.message,
        );
      case AuthenticatedApiFailureKind.malformedResponse:
        return AttendanceFailure(
          AttendanceFailureKind.malformedResponse,
          error.message,
        );
    }
  }
}
