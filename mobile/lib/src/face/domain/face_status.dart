enum FaceEnrollmentStatus {
  notEnrolled('NOT_ENROLLED', 'Belum terdaftar'),
  enrolled('ENROLLED', 'Terdaftar');

  const FaceEnrollmentStatus(this.value, this.label);

  final String value;
  final String label;

  static FaceEnrollmentStatus parse(Object? value) {
    if (value == enrolled.value) {
      return enrolled;
    }
    if (value == notEnrolled.value) {
      return notEnrolled;
    }
    throw const FormatException('Status wajah tidak valid.');
  }
}

class FaceStatus {
  const FaceStatus({
    required this.enrolled,
    required this.status,
    this.embeddingModel,
    this.embeddingVersion,
    this.enrolledAt,
  });

  final bool enrolled;
  final FaceEnrollmentStatus status;
  final String? embeddingModel;
  final String? embeddingVersion;
  final DateTime? enrolledAt;

  factory FaceStatus.fromJson(Object? raw) {
    if (raw is! Map) {
      throw const FormatException('Data status wajah bukan object.');
    }
    final json = Map<String, Object?>.from(raw);
    final enrolled = json['enrolled'];
    if (enrolled is! bool) {
      throw const FormatException('Field enrolled tidak valid.');
    }
    final status = FaceEnrollmentStatus.parse(json['face_status']);
    final enrolledAtRaw = json['enrolled_at'];
    return FaceStatus(
      enrolled: enrolled,
      status: status,
      embeddingModel: json['embedding_model'] as String?,
      embeddingVersion: json['embedding_version'] as String?,
      enrolledAt: enrolledAtRaw is String
          ? DateTime.parse(enrolledAtRaw)
          : null,
    );
  }
}
