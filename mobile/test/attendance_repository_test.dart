import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/attendance/data/attendance_api_client.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_failure.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_models.dart';
import 'package:r3_ti_faceattend/src/core/network/authenticated_api_client.dart';

import 'attendance_test_fakes.dart';

void main() {
  test('today berhasil', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(response(200, okTodayPayload()));
    final api = HttpAttendanceApiClient(client: requester);

    final today = await api.getTodayAttendance();

    expect(today.state, AttendanceState.notCheckedIn);
    expect(requester.lastPath, '/attendance/today');
  });

  test('check-in berhasil', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(
          201,
          okTodayPayload(
            state: 'CHECKED_IN',
            checkInAt: '2026-08-05T01:00:00Z',
            canCheckIn: false,
            canCheckOut: true,
          ),
        ),
      );
    final api = HttpAttendanceApiClient(client: requester);

    final today = await api.checkIn(locationPayload());

    expect(today.state, AttendanceState.checkedIn);
    expect(requester.lastMethod, 'POST');
    expect(requester.lastPath, '/attendance/check-in');
    expect(requester.lastBody, <String, Object?>{
      'latitude': -6.98946,
      'longitude': 110.416735,
      'accuracy_meters': 12.5,
    });
  });

  test('check-out berhasil', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(
          200,
          okTodayPayload(
            state: 'COMPLETED',
            checkInAt: '2026-08-05T01:00:00Z',
            checkOutAt: '2026-08-05T09:00:00Z',
            canCheckIn: false,
          ),
        ),
      );
    final api = HttpAttendanceApiClient(client: requester);

    final today = await api.checkOut(locationPayload());

    expect(today.state, AttendanceState.completed);
    expect(requester.lastPath, '/attendance/check-out');
  });

  test('401 dari authenticated client dipetakan session expired', () async {
    final requester = FakeAuthenticatedRequester()
      ..error = const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.getTodayAttendance(),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.sessionExpired,
        ),
      ),
    );
  });

  test('409 dipetakan menjadi pesan aman', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(409, <String, Object?>{
          'status': 'error',
          'message': 'pegawai sudah check-in',
        }),
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.checkIn(locationPayload()),
      throwsA(
        isA<AttendanceFailure>()
            .having(
              (error) => error.kind,
              'kind',
              AttendanceFailureKind.alreadyCheckedIn,
            )
            .having(
              (error) => error.message,
              'message',
              'Anda sudah melakukan check-in hari ini.',
            ),
      ),
    );
  });

  test('422 akurasi GPS dipetakan khusus', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(422, <String, Object?>{
          'status': 'error',
          'message': 'akurasi lokasi belum memenuhi batas',
        }),
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.checkIn(locationPayload()),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.poorAccuracy,
        ),
      ),
    );
  });

  test('403 luar radius dipetakan khusus', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(403, <String, Object?>{
          'status': 'error',
          'message': 'pegawai berada di luar radius lokasi kantor',
        }),
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.checkIn(locationPayload()),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.outsideGeofence,
        ),
      ),
    );
  });

  test('404 assignment lokasi dipetakan khusus', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(404, <String, Object?>{
          'status': 'error',
          'message': 'lokasi kerja belum ditugaskan',
        }),
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.checkIn(locationPayload()),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.locationAssignmentMissing,
        ),
      ),
    );
  });

  test('timeout dipetakan menjadi error aman', () async {
    final requester = FakeAuthenticatedRequester()
      ..error = const AuthenticatedApiFailure(
        AuthenticatedApiFailureKind.requestTimeout,
        'Request terlalu lama. Coba lagi.',
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.getTodayAttendance(),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.requestTimeout,
        ),
      ),
    );
  });

  test('history mengirim pagination', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(response(200, okHistoryPayload()));
    final api = HttpAttendanceApiClient(client: requester);

    await api.getAttendanceHistory(page: 2, pageSize: 20);

    expect(requester.lastQueryParameters, <String, String>{
      'page': '2',
      'page_size': '20',
    });
  });

  test('location requirement berhasil diparse', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(response(200, locationRequirementPayload()));
    final api = HttpAttendanceApiClient(client: requester);

    final requirement = await api.getLocationRequirement();

    expect(requester.lastMethod, 'GET');
    expect(requester.lastPath, '/attendance/location-requirement');
    expect(requirement.assignment.status, 'CURRENT');
    expect(requirement.officeLocation.name, contains('Kantor'));
  });

  test('location requirement gagal dipetakan aman', () async {
    final requester = FakeAuthenticatedRequester()
      ..responses.add(
        response(404, <String, Object?>{
          'status': 'error',
          'message': 'lokasi kerja belum ditugaskan',
        }),
      );
    final api = HttpAttendanceApiClient(client: requester);

    expect(
      api.getLocationRequirement(),
      throwsA(
        isA<AttendanceFailure>().having(
          (error) => error.kind,
          'kind',
          AttendanceFailureKind.locationAssignmentMissing,
        ),
      ),
    );
  });
}

AttendanceLocationPayload locationPayload() {
  return const AttendanceLocationPayload(
    latitude: -6.98946,
    longitude: 110.416735,
    accuracyMeters: 12.5,
  );
}
