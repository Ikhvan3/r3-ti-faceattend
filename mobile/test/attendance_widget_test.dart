import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:r3_ti_faceattend/src/attendance/data/attendance_repository.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_models.dart';
import 'package:r3_ti_faceattend/src/attendance/presentation/attendance_card.dart';
import 'package:r3_ti_faceattend/src/attendance/presentation/attendance_controller.dart';
import 'package:r3_ti_faceattend/src/attendance/presentation/attendance_history_page.dart';

import 'attendance_test_fakes.dart';

void main() {
  testWidgets('kartu NOT_CHECKED_IN menampilkan tombol Check-in', (
    tester,
  ) async {
    await tester.pumpWidget(await _attendanceCardApp(attendanceToday()));

    expect(find.text('Belum check-in'), findsOneWidget);
    expect(find.text('Check-in'), findsNWidgets(2));
  });

  testWidgets('kartu CHECKED_IN menampilkan tombol Check-out', (tester) async {
    await tester.pumpWidget(
      await _attendanceCardApp(
        attendanceToday(state: AttendanceState.checkedIn),
      ),
    );

    expect(find.text('Sudah check-in'), findsOneWidget);
    expect(find.text('Check-out'), findsNWidgets(2));
  });

  testWidgets('kartu COMPLETED tidak menampilkan tombol aktif', (tester) async {
    await tester.pumpWidget(
      await _attendanceCardApp(
        attendanceToday(state: AttendanceState.completed),
      ),
    );

    expect(find.text('Absensi selesai'), findsOneWidget);
  });

  testWidgets('loading state', (tester) async {
    final controller = AttendanceController(
      fakeAttendanceRepository(FakeAttendanceApi()),
      FakeLocationService(),
    );
    await tester.pumpWidget(_controllerApp(controller));

    expect(find.byType(LinearProgressIndicator), findsWidgets);
  });

  testWidgets('error state', (tester) async {
    final api = FakeAttendanceApi()..todayError = timeoutFailure;
    final controller = AttendanceController(
      fakeAttendanceRepository(api),
      FakeLocationService(),
    );
    await controller.initialize();

    await tester.pumpWidget(_controllerApp(controller));

    expect(find.text('Request terlalu lama. Coba lagi.'), findsOneWidget);
    expect(find.text('Coba Lagi'), findsOneWidget);
  });

  testWidgets('dialog konfirmasi check-in', (tester) async {
    final api = FakeAttendanceApi();
    final controller = AttendanceController(
      fakeAttendanceRepository(api),
      FakeLocationService(),
    );
    await controller.initialize();
    await tester.pumpWidget(_controllerApp(controller));

    await tester.tap(find.text('Check-in').last);
    await tester.pumpAndSettle();

    expect(find.text('Konfirmasi Check-in'), findsOneWidget);
    expect(
      find.text(
        'Anda akan melakukan check-in menggunakan lokasi perangkat dan waktu server.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('history empty state', (tester) async {
    await tester.pumpWidget(
      _historyApp(fakeAttendanceRepository(FakeAttendanceApi())),
    );
    await tester.pumpAndSettle();

    expect(find.text('Belum ada riwayat absensi.'), findsOneWidget);
  });

  testWidgets('history record tampil', (tester) async {
    final api = FakeAttendanceApi()
      ..historyResult = AttendanceHistoryResponse(
        items: <AttendanceRecord>[attendanceRecord()],
        pagination: const AttendancePagination(
          page: 1,
          pageSize: 10,
          totalItems: 1,
          totalPages: 1,
        ),
      );
    await tester.pumpWidget(_historyApp(fakeAttendanceRepository(api)));
    await tester.pumpAndSettle();

    expect(find.textContaining('Jadwal Kerja Dummy TI'), findsOneWidget);
    expect(find.textContaining('Check-in'), findsOneWidget);
  });
}

Future<Widget> _attendanceCardApp(AttendanceToday today) async {
  final controller = AttendanceController(
    fakeAttendanceRepository(FakeAttendanceApi()..todayResult = today),
    FakeLocationService(),
  );
  await controller.initialize();
  return _controllerApp(controller);
}

Widget _controllerApp(AttendanceController controller) {
  return MaterialApp(
    home: ChangeNotifierProvider<AttendanceController>.value(
      value: controller,
      child: const Scaffold(body: AttendanceCard()),
    ),
  );
}

Widget _historyApp(AttendanceRepository repository) {
  return MaterialApp(
    home: Provider<AttendanceRepository>.value(
      value: repository,
      child: const AttendanceHistoryPage(),
    ),
  );
}
