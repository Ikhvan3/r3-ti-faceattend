enum AttendanceFailureKind {
  sessionExpired,
  apiUnavailable,
  requestTimeout,
  locationServiceDisabled,
  locationPermissionDenied,
  locationPermissionDeniedForever,
  locationTimeout,
  poorAccuracy,
  locationAssignmentMissing,
  outsideGeofence,
  scheduleUnavailable,
  alreadyCheckedIn,
  notCheckedIn,
  alreadyCheckedOut,
  accountForbidden,
  malformedResponse,
  internalError,
}

class AttendanceFailure implements Exception {
  const AttendanceFailure(this.kind, this.message);

  final AttendanceFailureKind kind;
  final String message;

  @override
  String toString() => message;
}
