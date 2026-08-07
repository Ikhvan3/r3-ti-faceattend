import 'dart:async';

import 'package:geolocator/geolocator.dart' as geolocator;

enum AttendanceLocationPermission { denied, deniedForever, whileInUse, always }

class AttendancePosition {
  const AttendancePosition({
    required this.latitude,
    required this.longitude,
    required this.accuracyMeters,
  });

  final double latitude;
  final double longitude;
  final double accuracyMeters;
}

class AttendanceLocationServiceDisabledException implements Exception {
  const AttendanceLocationServiceDisabledException();
}

abstract class LocationService {
  Future<bool> isLocationServiceEnabled();
  Future<AttendanceLocationPermission> checkPermission();
  Future<AttendanceLocationPermission> requestPermission();
  Future<AttendancePosition> getCurrentPosition();
  Future<bool> openLocationSettings();
  Future<bool> openAppSettings();
}

class GeolocatorLocationService implements LocationService {
  const GeolocatorLocationService({this.timeout = const Duration(seconds: 12)});

  final Duration timeout;

  @override
  Future<AttendanceLocationPermission> checkPermission() async {
    return _mapPermission(await geolocator.Geolocator.checkPermission());
  }

  @override
  Future<AttendancePosition> getCurrentPosition() async {
    try {
      final position = await geolocator.Geolocator.getCurrentPosition(
        locationSettings: geolocator.LocationSettings(
          accuracy: geolocator.LocationAccuracy.high,
          timeLimit: timeout,
        ),
      );
      return AttendancePosition(
        latitude: position.latitude,
        longitude: position.longitude,
        accuracyMeters: position.accuracy,
      );
    } on TimeoutException {
      rethrow;
    } on geolocator.LocationServiceDisabledException {
      throw const AttendanceLocationServiceDisabledException();
    }
  }

  @override
  Future<bool> isLocationServiceEnabled() {
    return geolocator.Geolocator.isLocationServiceEnabled();
  }

  @override
  Future<bool> openAppSettings() {
    return geolocator.Geolocator.openAppSettings();
  }

  @override
  Future<bool> openLocationSettings() {
    return geolocator.Geolocator.openLocationSettings();
  }

  @override
  Future<AttendanceLocationPermission> requestPermission() async {
    return _mapPermission(await geolocator.Geolocator.requestPermission());
  }

  AttendanceLocationPermission _mapPermission(
    geolocator.LocationPermission permission,
  ) {
    switch (permission) {
      case geolocator.LocationPermission.denied:
        return AttendanceLocationPermission.denied;
      case geolocator.LocationPermission.deniedForever:
        return AttendanceLocationPermission.deniedForever;
      case geolocator.LocationPermission.whileInUse:
        return AttendanceLocationPermission.whileInUse;
      case geolocator.LocationPermission.always:
        return AttendanceLocationPermission.always;
      case geolocator.LocationPermission.unableToDetermine:
        return AttendanceLocationPermission.denied;
    }
  }
}
