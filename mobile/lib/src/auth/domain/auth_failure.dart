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

bool isAuthoritativeSessionFailure(AuthFailure error) {
  switch (error.kind) {
    case AuthFailureKind.invalidCredentials:
    case AuthFailureKind.sessionExpired:
    case AuthFailureKind.accountInactive:
    case AuthFailureKind.accountSuspended:
    case AuthFailureKind.forbiddenRole:
      return true;
    case AuthFailureKind.apiUnavailable:
    case AuthFailureKind.requestTimeout:
    case AuthFailureKind.invalidBackendResponse:
    case AuthFailureKind.internalError:
    case AuthFailureKind.storageError:
      return false;
  }
}

bool isTransientAuthFailure(AuthFailure error) {
  switch (error.kind) {
    case AuthFailureKind.apiUnavailable:
    case AuthFailureKind.requestTimeout:
    case AuthFailureKind.internalError:
      return true;
    case AuthFailureKind.invalidCredentials:
    case AuthFailureKind.sessionExpired:
    case AuthFailureKind.accountInactive:
    case AuthFailureKind.accountSuspended:
    case AuthFailureKind.forbiddenRole:
    case AuthFailureKind.invalidBackendResponse:
    case AuthFailureKind.storageError:
      return false;
  }
}
