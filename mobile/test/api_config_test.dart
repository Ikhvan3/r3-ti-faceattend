import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/config/api_config.dart';

void main() {
  test('base URL dengan /api/v1 tidak digandakan', () {
    const config = ApiConfig(baseUrl: 'http://10.0.2.2:8080/api/v1');

    expect(
      config.buildUrl('/auth/login'),
      'http://10.0.2.2:8080/api/v1/auth/login',
    );
  });

  test('leading slash endpoint tidak menghapus path base URL', () {
    const config = ApiConfig(baseUrl: 'http://127.0.0.1:8080/api/v1/');

    expect(config.buildUrl('/auth/me'), 'http://127.0.0.1:8080/api/v1/auth/me');
  });

  test('endpoint tanpa leading slash tetap valid', () {
    const config = ApiConfig(baseUrl: 'http://127.0.0.1:8080/api/v1');

    expect(
      config.buildUrl('auth/refresh'),
      'http://127.0.0.1:8080/api/v1/auth/refresh',
    );
  });
}
