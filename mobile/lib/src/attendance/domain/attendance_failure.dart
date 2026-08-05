enum AttendanceFailureKind {
  sessionExpired,
  apiUnavailable,
  requestTimeout,
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
