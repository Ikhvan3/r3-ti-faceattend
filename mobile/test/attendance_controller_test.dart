import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_failure.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_models.dart';
import 'package:r3_ti_faceattend/src/attendance/presentation/attendance_controller.dart';

import 'attendance_test_fakes.dart';

void main() {
  test('initialize menjadi loaded', () async {
    final api = FakeAttendanceApi();
    final controller = AttendanceController(fakeAttendanceRepository(api));

    await controller.initialize();

    expect(controller.status, AttendanceControllerStatus.loaded);
    expect(controller.today?.state, AttendanceState.notCheckedIn);
  });

  test('initialize gagal', () async {
    final api = FakeAttendanceApi()..todayError = timeoutFailure;
    final controller = AttendanceController(fakeAttendanceRepository(api));

    await controller.initialize();

    expect(controller.status, AttendanceControllerStatus.failure);
    expect(controller.errorMessage, 'Request terlalu lama. Coba lagi.');
  });

  test('check-in loading lalu berhasil dan refresh data', () async {
    final api = FakeAttendanceApi()
      ..actionCompleter = Completer<void>()
      ..todayResult = attendanceToday();
    final controller = AttendanceController(fakeAttendanceRepository(api));
    await controller.initialize();

    final future = controller.checkIn();
    expect(controller.status, AttendanceControllerStatus.actionLoading);
    api.todayResult = attendanceToday(state: AttendanceState.checkedIn);
    api.actionCompleter?.complete();
    await future;

    expect(controller.status, AttendanceControllerStatus.loaded);
    expect(controller.today?.state, AttendanceState.checkedIn);
    expect(api.todayCalls, 2);
  });

  test('check-in conflict tetap refresh data', () async {
    final api = FakeAttendanceApi()
      ..checkInError = alreadyCheckedInFailure
      ..todayResult = attendanceToday(state: AttendanceState.checkedIn);
    final controller = AttendanceController(fakeAttendanceRepository(api));
    await controller.initialize();

    await controller.checkIn();

    expect(controller.status, AttendanceControllerStatus.loaded);
    expect(controller.today?.state, AttendanceState.checkedIn);
    expect(controller.errorMessage, 'Anda sudah melakukan check-in hari ini.');
  });

  test('check-out berhasil', () async {
    final api = FakeAttendanceApi()
      ..todayResult = attendanceToday(state: AttendanceState.checkedIn);
    final controller = AttendanceController(fakeAttendanceRepository(api));
    await controller.initialize();
    api.todayResult = attendanceToday(state: AttendanceState.completed);

    await controller.checkOut();

    expect(controller.today?.state, AttendanceState.completed);
  });

  test('multiple action dicegah', () async {
    final api = FakeAttendanceApi()..actionCompleter = Completer<void>();
    final controller = AttendanceController(fakeAttendanceRepository(api));
    await controller.initialize();

    final first = controller.checkIn();
    final second = controller.checkOut();
    api.actionCompleter?.complete();
    await Future.wait(<Future<void>>[first, second]);

    expect(api.checkInCalls, 1);
    expect(api.checkOutCalls, 0);
  });

  test('refresh today', () async {
    final api = FakeAttendanceApi();
    final controller = AttendanceController(fakeAttendanceRepository(api));
    await controller.initialize();

    api.todayResult = attendanceToday(state: AttendanceState.checkedIn);
    await controller.refreshToday();

    expect(controller.today?.state, AttendanceState.checkedIn);
  });

  test('session expired ditandai', () async {
    final api = FakeAttendanceApi()
      ..todayError = const AttendanceFailure(
        AttendanceFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );
    final controller = AttendanceController(fakeAttendanceRepository(api));

    await controller.initialize();

    expect(controller.sessionExpired, isTrue);
  });
}
