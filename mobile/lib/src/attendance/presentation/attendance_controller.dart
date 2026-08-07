import 'dart:async';

import 'package:flutter/foundation.dart';

import '../data/attendance_repository.dart';
import '../data/location_service.dart';
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
  AttendanceController(this._repository, this._locationService);

  final AttendanceRepository _repository;
  final LocationService _locationService;

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
    Future<AttendanceToday> Function(AttendanceLocationPayload location)
    operation,
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
      final location = await _currentLocationPayload();
      _today = await operation(location);
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
      case AttendanceFailureKind.locationServiceDisabled:
      case AttendanceFailureKind.locationPermissionDenied:
      case AttendanceFailureKind.locationPermissionDeniedForever:
      case AttendanceFailureKind.locationTimeout:
      case AttendanceFailureKind.poorAccuracy:
      case AttendanceFailureKind.locationAssignmentMissing:
      case AttendanceFailureKind.outsideGeofence:
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

  Future<AttendanceLocationPayload> _currentLocationPayload() async {
    if (!await _locationService.isLocationServiceEnabled()) {
      throw const AttendanceFailure(
        AttendanceFailureKind.locationServiceDisabled,
        'Layanan lokasi belum aktif. Aktifkan GPS lalu coba lagi.',
      );
    }

    var permission = await _locationService.checkPermission();
    if (permission == AttendanceLocationPermission.denied) {
      permission = await _locationService.requestPermission();
    }

    switch (permission) {
      case AttendanceLocationPermission.whileInUse:
      case AttendanceLocationPermission.always:
        break;
      case AttendanceLocationPermission.denied:
        throw const AttendanceFailure(
          AttendanceFailureKind.locationPermissionDenied,
          'Izin lokasi diperlukan untuk melakukan absensi.',
        );
      case AttendanceLocationPermission.deniedForever:
        throw const AttendanceFailure(
          AttendanceFailureKind.locationPermissionDeniedForever,
          'Izin lokasi ditolak permanen. Buka pengaturan aplikasi untuk mengaktifkannya.',
        );
    }

    try {
      final position = await _locationService.getCurrentPosition();
      return AttendanceLocationPayload(
        latitude: position.latitude,
        longitude: position.longitude,
        accuracyMeters: position.accuracyMeters,
      );
    } on TimeoutException {
      throw const AttendanceFailure(
        AttendanceFailureKind.locationTimeout,
        'GPS belum mendapatkan lokasi terbaru. Coba lagi di area terbuka.',
      );
    } on AttendanceLocationServiceDisabledException {
      throw const AttendanceFailure(
        AttendanceFailureKind.locationServiceDisabled,
        'Layanan lokasi belum aktif. Aktifkan GPS lalu coba lagi.',
      );
    }
  }
}
