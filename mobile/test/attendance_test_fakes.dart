import 'dart:async';

import 'package:r3_ti_faceattend/src/attendance/data/attendance_api_client.dart';
import 'package:r3_ti_faceattend/src/attendance/data/attendance_repository.dart';
import 'package:r3_ti_faceattend/src/attendance/data/location_service.dart';
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
  Map<String, Object?>? lastBody;

  @override
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
    Map<String, Object?>? body,
  }) async {
    calls++;
    lastMethod = method;
    lastPath = path;
    lastQueryParameters = queryParameters;
    lastBody = body;
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
  LocationRequirement? locationRequirementResult;
  Object? todayError;
  Object? checkInError;
  Object? checkOutError;
  Object? historyError;
  Object? locationRequirementError;
  Completer<void>? actionCompleter;
  int todayCalls = 0;
  int checkInCalls = 0;
  int checkOutCalls = 0;
  int historyCalls = 0;
  int locationRequirementCalls = 0;
  AttendanceLocationPayload? lastLocation;

  @override
  Future<AttendanceToday> getTodayAttendance() async {
    todayCalls++;
    if (todayError != null) {
      throw todayError!;
    }
    return todayResult ?? attendanceToday();
  }

  @override
  Future<AttendanceToday> checkIn(
    AttendanceLocationPayload location, {
    required String verificationGrant,
  }) async {
    checkInCalls++;
    lastLocation = location;
    await actionCompleter?.future;
    if (checkInError != null) {
      throw checkInError!;
    }
    return checkInResult ?? attendanceToday(state: AttendanceState.checkedIn);
  }

  @override
  Future<AttendanceToday> checkOut(
    AttendanceLocationPayload location, {
    required String verificationGrant,
  }) async {
    checkOutCalls++;
    lastLocation = location;
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

  @override
  Future<LocationRequirement> getLocationRequirement() async {
    locationRequirementCalls++;
    if (locationRequirementError != null) {
      throw locationRequirementError!;
    }
    return locationRequirementResult ?? locationRequirement();
  }
}

AttendanceRepository fakeAttendanceRepository(FakeAttendanceApi api) {
  return AttendanceRepository(api: api);
}

class FakeLocationService implements LocationService {
  bool serviceEnabled = true;
  AttendanceLocationPermission permission =
      AttendanceLocationPermission.whileInUse;
  AttendanceLocationPermission requestedPermission =
      AttendanceLocationPermission.whileInUse;
  AttendancePosition position = const AttendancePosition(
    latitude: -6.98946,
    longitude: 110.416735,
    accuracyMeters: 12.5,
  );
  Object? positionError;
  int positionCalls = 0;
  int requestPermissionCalls = 0;

  @override
  Future<AttendanceLocationPermission> checkPermission() async {
    return permission;
  }

  @override
  Future<AttendancePosition> getCurrentPosition() async {
    positionCalls++;
    if (positionError != null) {
      throw positionError!;
    }
    return position;
  }

  @override
  Future<bool> isLocationServiceEnabled() async {
    return serviceEnabled;
  }

  @override
  Future<bool> openAppSettings() async {
    return true;
  }

  @override
  Future<bool> openLocationSettings() async {
    return true;
  }

  @override
  Future<AttendanceLocationPermission> requestPermission() async {
    requestPermissionCalls++;
    permission = requestedPermission;
    return requestedPermission;
  }
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
    checkInLocation: state == AttendanceState.notCheckedIn
        ? null
        : attendanceLocationEvidence(),
    checkOutLocation: state == AttendanceState.completed
        ? attendanceLocationEvidence()
        : null,
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
    checkInLocation: attendanceLocationEvidence(),
    checkOutLocation: state == AttendanceState.completed
        ? attendanceLocationEvidence()
        : null,
  );
}

AttendanceLocationEvidence attendanceLocationEvidence() {
  return const AttendanceLocationEvidence(
    officeLocationId: '00000000-0000-4000-8000-000000000020',
    officeLocationName: 'Kantor PTPN I Regional 3 Semarang',
    accuracyMeters: 12.5,
    distanceMeters: 0,
    insideGeofence: true,
  );
}

LocationRequirement locationRequirement() {
  return LocationRequirement(
    assignment: LocationAssignmentRequirement(
      id: '00000000-0000-4000-8000-000000000040',
      officeLocation: officeLocation(),
      effectiveFrom: '2026-08-06',
      effectiveTo: null,
      status: 'CURRENT',
    ),
    officeLocation: officeLocation(),
  );
}

OfficeLocation officeLocation() {
  return const OfficeLocation(
    id: '00000000-0000-4000-8000-000000000020',
    name: 'Kantor PTPN I Regional 3 Semarang',
    address: null,
    latitude: -6.98946,
    longitude: 110.416735,
    radiusMeters: 100,
    isActive: true,
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
      'check_in_location': checkInAt == null ? null : locationEvidenceJson(),
      'check_out_location': checkOutAt == null ? null : locationEvidenceJson(),
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
    'check_in_location': locationEvidenceJson(),
    'check_out_location': null,
  };
}

Map<String, Object?> locationEvidenceJson() {
  return <String, Object?>{
    'office_location_id': '00000000-0000-4000-8000-000000000020',
    'office_location_name': 'Kantor PTPN I Regional 3 Semarang',
    'accuracy_meters': 12.5,
    'distance_meters': 0,
    'inside_geofence': true,
  };
}

Map<String, Object?> locationRequirementPayload() {
  return <String, Object?>{
    'status': 'ok',
    'message': 'kebutuhan lokasi berhasil dibaca',
    'data': <String, Object?>{
      'assignment': <String, Object?>{
        'id': '00000000-0000-4000-8000-000000000040',
        'user': <String, Object?>{
          'id': '00000000-0000-4000-8000-000000000001',
          'employee_number': 'EMP-001',
          'name': 'Pegawai Dummy',
          'email': 'pegawai.dummy@example.test',
          'phone': null,
          'position': 'Staf TI',
          'role': 'USER',
          'account_status': 'ACTIVE',
        },
        'office_location': officeLocationJson(),
        'effective_from': '2026-08-06',
        'effective_to': null,
        'status': 'CURRENT',
        'created_at': '2026-08-06T01:00:00Z',
        'updated_at': '2026-08-06T01:00:00Z',
      },
      'office_location': officeLocationJson(),
    },
  };
}

Map<String, Object?> officeLocationJson() {
  return <String, Object?>{
    'id': '00000000-0000-4000-8000-000000000020',
    'name': 'Kantor PTPN I Regional 3 Semarang',
    'address': null,
    'latitude': -6.98946,
    'longitude': 110.416735,
    'radius_meters': 100,
    'is_active': true,
    'created_at': '2026-08-06T01:00:00Z',
    'updated_at': '2026-08-06T01:00:00Z',
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
