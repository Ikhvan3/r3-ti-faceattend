import 'dart:async';

import 'package:r3_ti_faceattend/src/attendance/data/attendance_api_client.dart';
import 'package:r3_ti_faceattend/src/attendance/data/attendance_repository.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_failure.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_models.dart';
import 'package:r3_ti_faceattend/src/core/network/authenticated_api_client.dart';

class FakeAuthenticatedRequester implements AuthenticatedRequester {
  final List<AuthenticatedApiResponse> responses = <AuthenticatedApiResponse>[];
  Object? error;
  int calls = 0;
  String? lastMethod;
  String? lastPath;
  Map<String, String>? lastQueryParameters;

  @override
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
  }) async {
    calls++;
    lastMethod = method;
    lastPath = path;
    lastQueryParameters = queryParameters;
    if (error != null) {
      throw error!;
    }
    if (responses.isEmpty) {
      throw StateError('No fake response configured.');
    }
    return responses.removeAt(0);
  }
}

class FakeAttendanceApi implements AttendanceApi {
  AttendanceToday? todayResult;
  AttendanceToday? checkInResult;
  AttendanceToday? checkOutResult;
  AttendanceHistoryResponse? historyResult;
  Object? todayError;
  Object? checkInError;
  Object? checkOutError;
  Object? historyError;
  Completer<void>? actionCompleter;
  int todayCalls = 0;
  int checkInCalls = 0;
  int checkOutCalls = 0;
  int historyCalls = 0;

  @override
  Future<AttendanceToday> getTodayAttendance() async {
    todayCalls++;
    if (todayError != null) {
      throw todayError!;
    }
    return todayResult ?? attendanceToday();
  }

  @override
  Future<AttendanceToday> checkIn() async {
    checkInCalls++;
    await actionCompleter?.future;
    if (checkInError != null) {
      throw checkInError!;
    }
    return checkInResult ?? attendanceToday(state: AttendanceState.checkedIn);
  }

  @override
  Future<AttendanceToday> checkOut() async {
    checkOutCalls++;
    await actionCompleter?.future;
    if (checkOutError != null) {
      throw checkOutError!;
    }
    return checkOutResult ?? attendanceToday(state: AttendanceState.completed);
  }

  @override
  Future<AttendanceHistoryResponse> getAttendanceHistory({
    required int page,
    required int pageSize,
  }) async {
    historyCalls++;
    if (historyError != null) {
      throw historyError!;
    }
    return historyResult ??
        const AttendanceHistoryResponse(
          items: <AttendanceRecord>[],
          pagination: AttendancePagination(
            page: 1,
            pageSize: 10,
            totalItems: 0,
            totalPages: 0,
          ),
        );
  }
}

AttendanceRepository fakeAttendanceRepository(FakeAttendanceApi api) {
  return AttendanceRepository(api: api);
}

AttendanceToday attendanceToday({
  AttendanceState state = AttendanceState.notCheckedIn,
  bool? canCheckIn,
  bool? canCheckOut,
}) {
  final checkInAt = state == AttendanceState.notCheckedIn
      ? null
      : DateTime.parse('2026-08-05T01:00:00Z').toLocal();
  final checkOutAt = state == AttendanceState.completed
      ? DateTime.parse('2026-08-05T09:00:00Z').toLocal()
      : null;

  return AttendanceToday(
    attendanceDate: DateTime(2026, 8, 5),
    schedule: workSchedule(),
    checkInAt: checkInAt,
    checkOutAt: checkOutAt,
    state: state,
    canCheckIn: canCheckIn ?? state == AttendanceState.notCheckedIn,
    canCheckOut: canCheckOut ?? state == AttendanceState.checkedIn,
  );
}

AttendanceRecord attendanceRecord({
  AttendanceState state = AttendanceState.checkedIn,
}) {
  return AttendanceRecord(
    id: '00000000-0000-4000-8000-000000000020',
    attendanceDate: DateTime(2026, 8, 5),
    schedule: workSchedule(),
    checkInAt: DateTime.parse('2026-08-05T01:00:00Z').toLocal(),
    checkOutAt: state == AttendanceState.completed
        ? DateTime.parse('2026-08-05T09:00:00Z').toLocal()
        : null,
    state: state,
  );
}

WorkSchedule workSchedule() {
  return const WorkSchedule(
    id: '00000000-0000-4000-8000-000000000010',
    name: 'Jadwal Kerja Dummy TI',
    startTime: '08:00',
    endTime: '17:00',
    graceMinutes: 15,
    isActive: true,
  );
}

Map<String, Object?> okTodayPayload({
  String state = 'NOT_CHECKED_IN',
  Object? checkInAt,
  Object? checkOutAt,
  bool canCheckIn = true,
  bool canCheckOut = false,
}) {
  return <String, Object?>{
    'status': 'ok',
    'message': 'absensi hari ini berhasil dibaca',
    'data': <String, Object?>{
      'attendance_date': '2026-08-05',
      'schedule': scheduleJson(),
      'check_in_at': checkInAt,
      'check_out_at': checkOutAt,
      'state': state,
      'can_check_in': canCheckIn,
      'can_check_out': canCheckOut,
    },
  };
}

Map<String, Object?> okHistoryPayload({
  List<Object?> items = const <Object?>[],
}) {
  return <String, Object?>{
    'status': 'ok',
    'message': 'riwayat absensi berhasil dibaca',
    'data': <String, Object?>{
      'items': items,
      'page': 1,
      'page_size': 10,
      'total_items': items.length,
      'total_pages': items.isEmpty ? 0 : 1,
    },
  };
}

Map<String, Object?> historyItemJson() {
  return <String, Object?>{
    'id': '00000000-0000-4000-8000-000000000020',
    'attendance_date': '2026-08-05',
    'schedule': scheduleJson(),
    'check_in_at': '2026-08-05T01:00:00Z',
    'check_out_at': null,
    'state': 'CHECKED_IN',
  };
}

Map<String, Object?> scheduleJson() {
  return <String, Object?>{
    'id': '00000000-0000-4000-8000-000000000010',
    'name': 'Jadwal Kerja Dummy TI',
    'start_time': '08:00',
    'end_time': '17:00',
    'grace_minutes': 15,
    'is_active': true,
  };
}

AuthenticatedApiResponse response(
  int statusCode,
  Map<String, Object?> payload,
) {
  return AuthenticatedApiResponse(statusCode: statusCode, payload: payload);
}

const alreadyCheckedInFailure = AttendanceFailure(
  AttendanceFailureKind.alreadyCheckedIn,
  'Anda sudah melakukan check-in hari ini.',
);

const timeoutFailure = AttendanceFailure(
  AttendanceFailureKind.requestTimeout,
  'Request terlalu lama. Coba lagi.',
);
