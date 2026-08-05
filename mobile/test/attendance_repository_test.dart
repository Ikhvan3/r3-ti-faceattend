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

    final today = await api.checkIn();

    expect(today.state, AttendanceState.checkedIn);
    expect(requester.lastMethod, 'POST');
    expect(requester.lastPath, '/attendance/check-in');
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

    final today = await api.checkOut();

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
      api.checkIn(),
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
}
