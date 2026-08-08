import 'dart:async';
import 'dart:math' as math;

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

enum AttendanceActionStep { location, face, submit }

typedef AttendanceVerificationGrantLoader = Future<String> Function();

class AttendanceController extends ChangeNotifier {
  AttendanceController(this._repository, this._locationService);

  final AttendanceRepository _repository;
  final LocationService _locationService;

  AttendanceControllerStatus _status = AttendanceControllerStatus.initial;
  AttendanceToday? _today;
  String? _errorMessage;
  AttendanceAction? _currentAction;
  AttendanceActionStep? _currentStep;
  bool _isRefreshing = false;
  bool _sessionExpired = false;

  AttendanceControllerStatus get status => _status;
  AttendanceToday? get today => _today;
  String? get errorMessage => _errorMessage;
  AttendanceAction? get currentAction => _currentAction;
  AttendanceActionStep? get currentStep => _currentStep;
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

  Future<void> checkIn(AttendanceVerificationGrantLoader loadGrant) async {
    await _runAction(AttendanceAction.checkIn, loadGrant, _repository.checkIn);
  }

  Future<void> checkOut(AttendanceVerificationGrantLoader loadGrant) async {
    await _runAction(
      AttendanceAction.checkOut,
      loadGrant,
      _repository.checkOut,
    );
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
    AttendanceVerificationGrantLoader loadGrant,
    Future<AttendanceToday> Function(
      AttendanceLocationPayload location, {
      required String verificationGrant,
    })
    operation,
  ) async {
    if (_status == AttendanceControllerStatus.actionLoading || _isRefreshing) {
      return;
    }

    _status = AttendanceControllerStatus.actionLoading;
    _currentAction = action;
    _currentStep = AttendanceActionStep.location;
    _errorMessage = null;
    _sessionExpired = false;
    notifyListeners();

    try {
      final location = await _currentLocationPayload();

      // Fast UX pre-check: do not open the camera when the current GPS point is
      // already outside the employee's assigned office radius. This is only an
      // early rejection. The backend still performs the authoritative geofence
      // validation again when check-in/check-out is submitted.
      await _ensureInsideAssignedGeofence(location);

      _currentStep = AttendanceActionStep.face;
      notifyListeners();
      final verificationGrant = await loadGrant();
      _currentStep = AttendanceActionStep.submit;
      notifyListeners();
      _today = await operation(location, verificationGrant: verificationGrant);
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
      _currentStep = null;
      notifyListeners();
    }
  }

  Future<void> _ensureInsideAssignedGeofence(
    AttendanceLocationPayload current,
  ) async {
    final requirement = await _repository.loadLocationRequirement();
    final office = requirement.officeLocation;
    if (!office.isActive) {
      throw const AttendanceFailure(
        AttendanceFailureKind.locationAssignmentMissing,
        'Lokasi kerja belum ditugaskan. Hubungi administrator.',
      );
    }

    final distance = _distanceMeters(
      current.latitude,
      current.longitude,
      office.latitude,
      office.longitude,
    );
    if (distance > office.radiusMeters) {
      throw const AttendanceFailure(
        AttendanceFailureKind.outsideGeofence,
        'Anda berada di luar area absensi.',
      );
    }
  }

  double _distanceMeters(
    double latitudeA,
    double longitudeA,
    double latitudeB,
    double longitudeB,
  ) {
    const earthRadiusMeters = 6371000.0;
    final lat1 = _degreesToRadians(latitudeA);
    final lat2 = _degreesToRadians(latitudeB);
    final deltaLat = _degreesToRadians(latitudeB - latitudeA);
    final deltaLon = _degreesToRadians(longitudeB - longitudeA);

    final haversine =
        math.sin(deltaLat / 2) * math.sin(deltaLat / 2) +
        math.cos(lat1) *
            math.cos(lat2) *
            math.sin(deltaLon / 2) *
            math.sin(deltaLon / 2);
    final angularDistance = 2 * math.atan2(math.sqrt(haversine), math.sqrt(1 - haversine));
    return earthRadiusMeters * angularDistance;
  }

  double _degreesToRadians(double degrees) => degrees * math.pi / 180;

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
      case AttendanceFailureKind.faceVerificationRejected:
      case AttendanceFailureKind.faceVerificationExpired:
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
        'Lokasi belum dapat diperoleh.',
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
        'Lokasi belum dapat diperoleh.',
      );
    } on AttendanceLocationServiceDisabledException {
      throw const AttendanceFailure(
        AttendanceFailureKind.locationServiceDisabled,
        'Lokasi belum dapat diperoleh.',
      );
    }
  }
}
