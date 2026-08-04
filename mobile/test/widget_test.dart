import 'package:flutter_test/flutter_test.dart';

import 'package:r3_ti_faceattend/main.dart';

void main() {
  testWidgets('shows application name', (WidgetTester tester) async {
    await tester.pumpWidget(const R3TiFaceAttendApp());

    expect(find.text('R3 TI FaceAttend'), findsOneWidget);
  });
}
