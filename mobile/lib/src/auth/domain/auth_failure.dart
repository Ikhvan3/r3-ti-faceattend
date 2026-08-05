enum AuthFailureKind {
  invalidCredentials,
  sessionExpired,
  accountInactive,
  accountSuspended,
  forbiddenRole,
  apiUnavailable,
  requestTimeout,
  invalidBackendResponse,
  internalError,
  storageError,
}

class AuthFailure implements Exception {
  const AuthFailure(this.kind, this.message);

  final AuthFailureKind kind;
  final String message;

  @override
  String toString() => message;
}
