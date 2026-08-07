import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:r3_ti_faceattend/src/face/presentation/face_enrollment_controller.dart';
import 'package:r3_ti_faceattend/src/face/presentation/face_status_card.dart';

import 'face_controller_test.dart';
import 'face_test_fakes.dart';

void main() {
  testWidgets('Home face card shows not enrolled state', (tester) async {
    await tester.pumpWidget(
      faceCardApp(FakeFaceApi(status: notEnrolledStatus)),
    );
    await tester.pump();

    expect(find.text('Belum terdaftar'), findsOneWidget);
    expect(find.text('Daftarkan Wajah'), findsOneWidget);
    expect(find.text('Uji Verifikasi Wajah'), findsNothing);
  });

  testWidgets('Home face card shows enrolled state', (tester) async {
    await tester.pumpWidget(faceCardApp(FakeFaceApi(status: enrolledStatus)));
    await tester.pump();

    expect(find.text('Terdaftar'), findsOneWidget);
    expect(find.text('Uji Verifikasi Wajah'), findsOneWidget);
    expect(find.text('Atur Ulang Wajah'), findsOneWidget);
  });

  testWidgets('reset confirmation cancel does not reset', (tester) async {
    final api = FakeFaceApi(status: enrolledStatus);
    await tester.pumpWidget(faceCardApp(api));
    await tester.pump();

    await tester.tap(find.text('Atur Ulang Wajah'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Batal'));
    await tester.pumpAndSettle();

    expect(api.resetCount, 0);
  });
}

Widget faceCardApp(FakeFaceApi api) {
  final controller = newFaceController(api: api)..loadStatus();
  return MaterialApp(
    home: ChangeNotifierProvider<FaceEnrollmentController>.value(
      value: controller,
      child: const Scaffold(body: FaceStatusCard()),
    ),
  );
}
