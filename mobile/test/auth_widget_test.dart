import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/main.dart';
import 'package:r3_ti_faceattend/src/attendance/domain/attendance_failure.dart';
import 'package:r3_ti_faceattend/src/auth/data/auth_repository.dart';

import 'attendance_test_fakes.dart';
import 'auth_test_fakes.dart';

void main() {
  testWidgets('login form tampil', (tester) async {
    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('R3 TI FaceAttend'), findsOneWidget);
    expect(find.text('Pegawai Divisi Teknologi Informasi'), findsOneWidget);
    expect(find.byType(TextFormField), findsNWidgets(2));
  });

  testWidgets('validasi field kosong', (tester) async {
    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Masuk'));
    await tester.pump();

    expect(find.text('Email wajib diisi.'), findsOneWidget);
    expect(find.text('Password wajib diisi.'), findsOneWidget);
  });

  testWidgets('loading button', (tester) async {
    final api = FakeAuthApi()..loginCompleter = Completer<void>();
    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: api,
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'password');
    await tester.tap(find.text('Masuk'));
    await tester.pump();

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    api.loginCompleter?.complete();
    await tester.pumpAndSettle();
  });

  testWidgets('error message', (tester) async {
    final api = FakeAuthApi()..loginError = invalidCredentialFailure;
    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: api,
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'wrong');
    await tester.tap(find.text('Masuk'));
    await tester.pumpAndSettle();

    expect(find.text('Email atau password tidak valid.'), findsOneWidget);
  });

  testWidgets('home menampilkan profil', (tester) async {
    final api = FakeAuthApi()..meResult = userProfile(name: 'Pegawai Dummy UI');
    final storage = FakeTokenStorage()
      ..accessToken = 'access-token'
      ..refreshToken = 'refresh-token';

    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(api: api, tokenStorage: storage),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Halo, Pegawai Dummy UI'), findsOneWidget);
    expect(find.text('EMP-001'), findsOneWidget);
    expect(find.text('Staf TI'), findsOneWidget);
  });

  testWidgets('password tidak tampil secara default', (tester) async {
    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(FakeAttendanceApi()),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    final passwordInput = tester.widget<EditableText>(
      find.byType(EditableText).last,
    );
    expect(passwordInput.obscureText, isTrue);
  });

  testWidgets('login tetap membuka Home ketika attendance today gagal', (
    tester,
  ) async {
    final authApi = FakeAuthApi();
    final storage = FakeTokenStorage();
    final attendanceApi = FakeAttendanceApi()..todayError = timeoutFailure;
    final locationService = FakeLocationService();

    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(api: authApi, tokenStorage: storage),
        attendanceRepository: fakeAttendanceRepository(attendanceApi),
        locationService: locationService,
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'password');
    await tester.tap(find.text('Masuk'));
    await tester.pumpAndSettle();

    expect(find.text('Halo, Pegawai Dummy'), findsOneWidget);
    expect(find.text('R3 TI FaceAttend'), findsNothing);
    expect(find.text('Request terlalu lama. Coba lagi.'), findsOneWidget);
    expect(storage.accessToken, 'access-token');
    expect(locationService.positionCalls, 0);
  });

  testWidgets('auth tidak dihapus karena attendance session expired', (
    tester,
  ) async {
    final storage = FakeTokenStorage();
    final attendanceApi = FakeAttendanceApi()
      ..todayError = const AttendanceFailure(
        AttendanceFailureKind.sessionExpired,
        'Session berakhir. Silakan login ulang.',
      );

    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: storage,
        ),
        attendanceRepository: fakeAttendanceRepository(attendanceApi),
        locationService: FakeLocationService(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'password');
    await tester.tap(find.text('Masuk'));
    await tester.pumpAndSettle();

    expect(find.text('Halo, Pegawai Dummy'), findsOneWidget);
    expect(find.text('Session berakhir. Silakan login ulang.'), findsOneWidget);
    expect(storage.accessToken, 'access-token');
  });

  testWidgets('login tetap membuka Home ketika location requirement gagal', (
    tester,
  ) async {
    final attendanceApi = FakeAttendanceApi()
      ..locationRequirementError = timeoutFailure;
    final locationService = FakeLocationService();

    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(attendanceApi),
        locationService: locationService,
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'password');
    await tester.tap(find.text('Masuk'));
    await tester.pumpAndSettle();

    expect(find.text('Halo, Pegawai Dummy'), findsOneWidget);
    expect(attendanceApi.locationRequirementCalls, 0);
    expect(locationService.positionCalls, 0);
  });

  testWidgets('GPS baru dipanggil ketika Check-in ditekan', (tester) async {
    tester.view.physicalSize = const Size(800, 1000);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final attendanceApi = FakeAttendanceApi();
    final locationService = FakeLocationService();

    await tester.pumpWidget(
      R3TiFaceAttendApp(
        authRepository: AuthRepository(
          api: FakeAuthApi(),
          tokenStorage: FakeTokenStorage(),
        ),
        attendanceRepository: fakeAttendanceRepository(attendanceApi),
        locationService: locationService,
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      'pegawai@example.test',
    );
    await tester.enterText(find.byType(TextFormField).last, 'password');
    await tester.tap(find.text('Masuk'));
    await tester.pumpAndSettle();

    expect(locationService.positionCalls, 0);

    final checkInButton = find.text('Check-in').last;
    await tester.tap(checkInButton);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Check-in').last);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(locationService.positionCalls, 1);
    expect(attendanceApi.checkInCalls, 0);
  });
}
