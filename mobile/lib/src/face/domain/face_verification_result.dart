class FaceVerificationResult {
  const FaceVerificationResult({required this.verified});

  final bool verified;

  factory FaceVerificationResult.fromJson(Object? value) {
    if (value is! Map<String, Object?>) {
      throw const FormatException('Invalid face verification response');
    }
    final verified = value['verified'];
    if (verified is! bool) {
      throw const FormatException('Invalid face verification result');
    }
    return FaceVerificationResult(verified: verified);
  }
}
