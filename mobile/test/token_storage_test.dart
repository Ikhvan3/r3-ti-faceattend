import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/auth/data/token_storage.dart';
import 'package:r3_ti_faceattend/src/auth/domain/auth_failure.dart';

import 'auth_test_fakes.dart';

void main() {
  test('secure token storage menyimpan dan membersihkan token', () async {
    final adapter = FakeSecureStorageAdapter();
    final storage = SecureTokenStorage(storage: adapter);

    await storage.saveTokens(userTokens());

    expect(await storage.readAccessToken(), 'access-token');
    expect(await storage.readRefreshToken(), 'refresh-token');

    await storage.clearTokens();

    expect(await storage.readAccessToken(), isNull);
    expect(await storage.readRefreshToken(), isNull);
  });

  test('secure token storage tidak mengekspos error internal', () async {
    final storage = SecureTokenStorage(
      storage: FakeSecureStorageAdapter()..throwOnWrite = true,
    );

    expect(() => storage.saveTokens(userTokens()), throwsA(isA<AuthFailure>()));
  });
}
