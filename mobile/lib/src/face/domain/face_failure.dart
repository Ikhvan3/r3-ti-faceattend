enum FaceFailureKind {
  cameraPermissionDenied,
  cameraUnavailable,
  noFace,
  multipleFaces,
  faceTooSmall,
  faceTooCloseToEdge,
  invalidPose,
  corruptInput,
  invalidEmbedding,
  duplicateEnrollment,
  notEnrolled,
  verificationRejected,
  accountForbidden,
  sessionExpired,
  apiUnavailable,
  requestTimeout,
  malformedResponse,
  internalError,
}

class FaceFailure implements Exception {
  const FaceFailure(this.kind, this.message);

  final FaceFailureKind kind;
  final String message;

  @override
  String toString() => message;
}
