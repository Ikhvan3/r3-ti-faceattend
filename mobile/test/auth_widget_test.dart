import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/main.dart';
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
      ),
    );
    await tester.pumpAndSettle();

    final passwordInput = tester.widget<EditableText>(
      find.byType(EditableText).last,
    );
    expect(passwordInput.obscureText, isTrue);
  });
}
