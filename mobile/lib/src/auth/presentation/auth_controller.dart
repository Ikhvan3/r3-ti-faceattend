import 'package:flutter/foundation.dart';

import '../data/auth_repository.dart';
import '../domain/auth_failure.dart';
import '../domain/user_profile.dart';

enum AuthControllerStatus {
  initial,
  loading,
  authenticated,
  unauthenticated,
  failure,
}

class AuthController extends ChangeNotifier {
  AuthController(this._repository);

  final AuthRepository _repository;

  AuthControllerStatus _status = AuthControllerStatus.initial;
  UserProfile? _user;
  String? _errorMessage;
  bool _isSubmitting = false;

  AuthControllerStatus get status => _status;
  UserProfile? get user => _user;
  String? get errorMessage => _errorMessage;
  bool get isSubmitting => _isSubmitting;

  Future<void> initialize() async {
    if (_status == AuthControllerStatus.loading) {
      return;
    }

    _setState(status: AuthControllerStatus.loading, errorMessage: null);
    final restoredUser = await _repository.restoreSession();
    if (restoredUser == null) {
      _setState(
        status: AuthControllerStatus.unauthenticated,
        user: null,
        errorMessage: null,
      );
      return;
    }

    _setState(
      status: AuthControllerStatus.authenticated,
      user: restoredUser,
      errorMessage: null,
    );
  }

  Future<void> login({required String email, required String password}) async {
    if (_isSubmitting) {
      return;
    }

    _isSubmitting = true;
    _setState(status: AuthControllerStatus.loading, errorMessage: null);

    try {
      final loggedInUser = await _repository.login(
        email: email,
        password: password,
      );
      _isSubmitting = false;
      _setState(
        status: AuthControllerStatus.authenticated,
        user: loggedInUser,
        errorMessage: null,
      );
    } on AuthFailure catch (error) {
      _isSubmitting = false;
      _setState(
        status: AuthControllerStatus.failure,
        user: null,
        errorMessage: error.message,
      );
    }
  }

  Future<void> logout() async {
    if (_isSubmitting) {
      return;
    }

    _isSubmitting = true;
    _setState(status: AuthControllerStatus.loading, errorMessage: null);
    await _repository.logout();
    _isSubmitting = false;
    _setState(
      status: AuthControllerStatus.unauthenticated,
      user: null,
      errorMessage: null,
    );
  }

  void clearError() {
    if (_errorMessage == null) {
      return;
    }

    _errorMessage = null;
    if (_status == AuthControllerStatus.failure) {
      _status = AuthControllerStatus.unauthenticated;
    }
    notifyListeners();
  }

  void _setState({
    required AuthControllerStatus status,
    UserProfile? user,
    String? errorMessage,
  }) {
    _status = status;
    _user = user;
    _errorMessage = errorMessage;
    notifyListeners();
  }
}
