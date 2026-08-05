import 'package:flutter/foundation.dart';

import '../data/attendance_repository.dart';
import '../domain/attendance_failure.dart';
import '../domain/attendance_models.dart';

enum AttendanceControllerStatus {
  initial,
  loading,
  loaded,
  actionLoading,
  failure,
}

enum AttendanceAction { checkIn, checkOut, refresh }

class AttendanceController extends ChangeNotifier {
  AttendanceController(this._repository);

  final AttendanceRepository _repository;

  AttendanceControllerStatus _status = AttendanceControllerStatus.initial;
  AttendanceToday? _today;
  String? _errorMessage;
  AttendanceAction? _currentAction;
  bool _isRefreshing = false;
  bool _sessionExpired = false;

  AttendanceControllerStatus get status => _status;
  AttendanceToday? get today => _today;
  String? get errorMessage => _errorMessage;
  AttendanceAction? get currentAction => _currentAction;
  bool get isRefreshing => _isRefreshing;
  bool get sessionExpired => _sessionExpired;
  bool get isBusy =>
      _status == AttendanceControllerStatus.loading ||
      _status == AttendanceControllerStatus.actionLoading ||
      _isRefreshing;

  Future<void> initialize() async {
    if (_status != AttendanceControllerStatus.initial) {
      return;
    }
    await _load(showInitialLoading: true);
  }

  Future<void> refreshToday() async {
    if (_isRefreshing || _status == AttendanceControllerStatus.actionLoading) {
      return;
    }
    await _load(showInitialLoading: _today == null, refreshing: _today != null);
  }

  Future<void> checkIn() async {
    await _runAction(AttendanceAction.checkIn, _repository.checkIn);
  }

  Future<void> checkOut() async {
    await _runAction(AttendanceAction.checkOut, _repository.checkOut);
  }

  void clearError() {
    if (_errorMessage == null && !_sessionExpired) {
      return;
    }
    _errorMessage = null;
    _sessionExpired = false;
    if (_status == AttendanceControllerStatus.failure && _today != null) {
      _status = AttendanceControllerStatus.loaded;
    }
    notifyListeners();
  }

  Future<void> _load({
    required bool showInitialLoading,
    bool refreshing = false,
  }) async {
    if (showInitialLoading) {
      _status = AttendanceControllerStatus.loading;
    }
    _isRefreshing = refreshing;
    _errorMessage = null;
    _sessionExpired = false;
    notifyListeners();

    try {
      _today = await _repository.loadToday();
      _status = AttendanceControllerStatus.loaded;
    } on AttendanceFailure catch (error) {
      _applyFailure(error);
    } finally {
      _isRefreshing = false;
      notifyListeners();
    }
  }

  Future<void> _runAction(
    AttendanceAction action,
    Future<AttendanceToday> Function() operation,
  ) async {
    if (_status == AttendanceControllerStatus.actionLoading || _isRefreshing) {
      return;
    }

    _status = AttendanceControllerStatus.actionLoading;
    _currentAction = action;
    _errorMessage = null;
    _sessionExpired = false;
    notifyListeners();

    try {
      _today = await operation();
      _today = await _repository.loadToday();
      _status = AttendanceControllerStatus.loaded;
    } on AttendanceFailure catch (error) {
      _applyFailure(error);
      if (_shouldRefreshAfter(error)) {
        try {
          _today = await _repository.loadToday();
          _status = AttendanceControllerStatus.loaded;
        } on AttendanceFailure catch (refreshError) {
          _applyFailure(refreshError);
        }
      }
    } finally {
      _currentAction = null;
      notifyListeners();
    }
  }

  bool _shouldRefreshAfter(AttendanceFailure error) {
    switch (error.kind) {
      case AttendanceFailureKind.alreadyCheckedIn:
      case AttendanceFailureKind.notCheckedIn:
      case AttendanceFailureKind.alreadyCheckedOut:
        return true;
      case AttendanceFailureKind.sessionExpired:
      case AttendanceFailureKind.apiUnavailable:
      case AttendanceFailureKind.requestTimeout:
      case AttendanceFailureKind.scheduleUnavailable:
      case AttendanceFailureKind.accountForbidden:
      case AttendanceFailureKind.malformedResponse:
      case AttendanceFailureKind.internalError:
        return false;
    }
  }

  void _applyFailure(AttendanceFailure error) {
    _status = AttendanceControllerStatus.failure;
    _errorMessage = error.message;
    _sessionExpired = error.kind == AttendanceFailureKind.sessionExpired;
  }
}
