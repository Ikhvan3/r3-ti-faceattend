import 'user_profile.dart';

class AuthTokenData {
  const AuthTokenData({
    required this.accessToken,
    required this.refreshToken,
    required this.tokenType,
    required this.expiresIn,
    required this.user,
  });

  factory AuthTokenData.fromJson(Object? value) {
    if (value is! Map<String, Object?>) {
      throw const FormatException('Data token tidak valid.');
    }

    final accessToken = value['access_token'];
    final refreshToken = value['refresh_token'];
    final tokenType = value['token_type'];
    final expiresIn = value['expires_in'];

    if (accessToken is! String || accessToken.isEmpty) {
      throw const FormatException('Access token tidak valid.');
    }
    if (refreshToken is! String || refreshToken.isEmpty) {
      throw const FormatException('Refresh token tidak valid.');
    }
    if (tokenType != 'Bearer') {
      throw const FormatException('Tipe token tidak valid.');
    }
    if (expiresIn is! int || expiresIn < 1) {
      throw const FormatException('Masa berlaku token tidak valid.');
    }

    return AuthTokenData(
      accessToken: accessToken,
      refreshToken: refreshToken,
      tokenType: tokenType as String,
      expiresIn: expiresIn,
      user: UserProfile.fromJson(value['user']),
    );
  }

  final String accessToken;
  final String refreshToken;
  final String tokenType;
  final int expiresIn;
  final UserProfile user;
}
