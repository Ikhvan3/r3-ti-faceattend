class FaceAttendanceGrant {
  const FaceAttendanceGrant({
    required this.verificationGrant,
    required this.expiresAt,
  });

  factory FaceAttendanceGrant.fromJson(Object? value) {
    if (value is! Map) {
      throw const FormatException('Grant verifikasi wajah tidak valid.');
    }
    final json = Map<String, Object?>.from(value);
    final grant = json['verification_grant'];
    final expiresAt = json['expires_at'];
    if (grant is! String || grant.trim().isEmpty) {
      throw const FormatException('Grant verifikasi wajah tidak valid.');
    }
    if (expiresAt is! String) {
      throw const FormatException('Masa berlaku verifikasi wajah tidak valid.');
    }
    final parsedExpiresAt = DateTime.tryParse(expiresAt);
    if (parsedExpiresAt == null) {
      throw const FormatException('Masa berlaku verifikasi wajah tidak valid.');
    }

    return FaceAttendanceGrant(
      verificationGrant: grant,
      expiresAt: parsedExpiresAt.toLocal(),
    );
  }

  final String verificationGrant;
  final DateTime expiresAt;
}
