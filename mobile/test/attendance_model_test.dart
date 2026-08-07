import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_models.dart';

import 'attendance_test_fakes.dart';

void main() {
  test('parse NOT_CHECKED_IN', () {
    final today = AttendanceToday.fromJson(okTodayPayload()['data']);

    expect(today.state, AttendanceState.notCheckedIn);
    expect(today.checkInAt, isNull);
    expect(today.checkOutAt, isNull);
    expect(today.checkInLocation, isNull);
    expect(today.checkOutLocation, isNull);
    expect(today.canCheckIn, isTrue);
  });

  test('parse CHECKED_IN', () {
    final today = AttendanceToday.fromJson(
      okTodayPayload(
        state: 'CHECKED_IN',
        checkInAt: '2026-08-05T01:00:00Z',
        canCheckIn: false,
        canCheckOut: true,
      )['data'],
    );

    expect(today.state, AttendanceState.checkedIn);
    expect(today.checkInAt, isNotNull);
    expect(today.checkOutAt, isNull);
    expect(today.canCheckOut, isTrue);
    expect(today.checkInLocation?.officeLocationName, contains('Kantor'));
    expect(today.checkInLocation?.accuracyMeters, 12.5);
  });

  test('parse COMPLETED', () {
    final today = AttendanceToday.fromJson(
      okTodayPayload(
        state: 'COMPLETED',
        checkInAt: '2026-08-05T01:00:00Z',
        checkOutAt: '2026-08-05T09:00:00Z',
        canCheckIn: false,
      )['data'],
    );

    expect(today.state, AttendanceState.completed);
    expect(today.checkOutAt, isNotNull);
  });

  test('timestamp nullable valid', () {
    final record = AttendanceRecord.fromJson(historyItemJson());

    expect(record.checkInAt, isNotNull);
    expect(record.checkOutAt, isNull);
  });

  test('malformed response ditolak', () {
    expect(
      () => AttendanceToday.fromJson(<String, Object?>{
        'attendance_date': '2026-08-05',
        'schedule': scheduleJson(),
        'state': 'UNKNOWN',
        'can_check_in': true,
        'can_check_out': false,
      }),
      throwsA(isA<FormatException>()),
    );
  });

  test('malformed location evidence ditolak', () {
    final payload = okTodayPayload(
      state: 'CHECKED_IN',
      checkInAt: '2026-08-05T01:00:00Z',
      canCheckIn: false,
      canCheckOut: true,
    );
    final data = Map<String, Object?>.from(payload['data']! as Map);
    data['check_in_location'] = <String, Object?>{
      'office_location_id': 'office-id',
      'office_location_name': 'Kantor',
      'accuracy_meters': '12.5',
      'distance_meters': 0,
      'inside_geofence': true,
    };

    expect(
      () => AttendanceToday.fromJson(data),
      throwsA(isA<FormatException>()),
    );
  });

  test('history kosong', () {
    final history = AttendanceHistoryResponse.fromJson(
      okHistoryPayload()['data'],
    );

    expect(history.items, isEmpty);
    expect(history.pagination.totalItems, 0);
  });

  test('pagination valid', () {
    final history = AttendanceHistoryResponse.fromJson(
      okHistoryPayload(items: <Object?>[historyItemJson()])['data'],
    );

    expect(history.items, hasLength(1));
    expect(history.pagination.page, 1);
    expect(history.pagination.totalPages, 1);
  });

  test('location requirement response valid', () {
    final requirement = LocationRequirement.fromJson(
      locationRequirementPayload()['data'],
    );

    expect(requirement.assignment.id, isNotEmpty);
    expect(
      requirement.assignment.officeLocation.id,
      requirement.officeLocation.id,
    );
    expect(requirement.assignment.status, 'CURRENT');
    expect(requirement.officeLocation.address, isNull);
    expect(requirement.officeLocation.latitude, -6.98946);
  });
}
