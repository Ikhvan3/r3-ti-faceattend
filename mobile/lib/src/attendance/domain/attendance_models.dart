enum AttendanceState {
  notCheckedIn,
  checkedIn,
  completed;

  factory AttendanceState.fromJson(Object? value) {
    switch (value) {
      case 'NOT_CHECKED_IN':
        return AttendanceState.notCheckedIn;
      case 'CHECKED_IN':
        return AttendanceState.checkedIn;
      case 'COMPLETED':
        return AttendanceState.completed;
      default:
        throw const FormatException('Status absensi tidak valid.');
    }
  }

  String get label {
    switch (this) {
      case AttendanceState.notCheckedIn:
        return 'Belum check-in';
      case AttendanceState.checkedIn:
        return 'Sudah check-in';
      case AttendanceState.completed:
        return 'Selesai';
    }
  }
}

class WorkSchedule {
  const WorkSchedule({
    required this.id,
    required this.name,
    required this.startTime,
    required this.endTime,
    required this.graceMinutes,
    required this.isActive,
  });

  factory WorkSchedule.fromJson(Object? value) {
    final json = _object(value, 'Jadwal kerja tidak valid.');
    final id = _string(json['id'], 'ID jadwal tidak valid.');
    final name = _string(json['name'], 'Nama jadwal tidak valid.');
    final startTime = _string(json['start_time'], 'Jam mulai tidak valid.');
    final endTime = _string(json['end_time'], 'Jam selesai tidak valid.');
    final graceMinutes = json['grace_minutes'];
    final isActive = json['is_active'];

    if (graceMinutes is! int || graceMinutes < 0) {
      throw const FormatException('Grace period tidak valid.');
    }
    if (isActive is! bool) {
      throw const FormatException('Status jadwal tidak valid.');
    }

    return WorkSchedule(
      id: id,
      name: name,
      startTime: startTime,
      endTime: endTime,
      graceMinutes: graceMinutes,
      isActive: isActive,
    );
  }

  final String id;
  final String name;
  final String startTime;
  final String endTime;
  final int graceMinutes;
  final bool isActive;
}

class AttendanceLocationPayload {
  const AttendanceLocationPayload({
    required this.latitude,
    required this.longitude,
    required this.accuracyMeters,
  });

  final double latitude;
  final double longitude;
  final double accuracyMeters;

  Map<String, Object?> toJson({required String verificationGrant}) {
    return <String, Object?>{
      'latitude': latitude,
      'longitude': longitude,
      'accuracy_meters': accuracyMeters,
      'verification_grant': verificationGrant,
    };
  }
}

class AttendanceLocationEvidence {
  const AttendanceLocationEvidence({
    required this.officeLocationId,
    required this.officeLocationName,
    required this.accuracyMeters,
    required this.distanceMeters,
    required this.insideGeofence,
  });

  factory AttendanceLocationEvidence.fromJson(Object? value) {
    if (value == null) {
      throw const FormatException('Evidence lokasi tidak valid.');
    }
    final json = _object(value, 'Evidence lokasi tidak valid.');
    return AttendanceLocationEvidence(
      officeLocationId: _string(
        json['office_location_id'],
        'ID lokasi kantor tidak valid.',
      ),
      officeLocationName: _string(
        json['office_location_name'],
        'Nama lokasi kantor tidak valid.',
      ),
      accuracyMeters: _number(
        json['accuracy_meters'],
        'Akurasi lokasi tidak valid.',
      ),
      distanceMeters: _number(
        json['distance_meters'],
        'Jarak lokasi tidak valid.',
      ),
      insideGeofence: _bool(
        json['inside_geofence'],
        'Status geofence tidak valid.',
      ),
    );
  }

  final String officeLocationId;
  final String officeLocationName;
  final double accuracyMeters;
  final double distanceMeters;
  final bool insideGeofence;
}

class OfficeLocation {
  const OfficeLocation({
    required this.id,
    required this.name,
    required this.address,
    required this.latitude,
    required this.longitude,
    required this.radiusMeters,
    required this.isActive,
  });

  factory OfficeLocation.fromJson(Object? value) {
    final json = _object(value, 'Lokasi kantor tidak valid.');
    final address = json['address'];
    if (address != null && address is! String) {
      throw const FormatException('Alamat lokasi kantor tidak valid.');
    }
    final radiusMeters = json['radius_meters'];
    final isActive = json['is_active'];
    if (radiusMeters is! int || radiusMeters <= 0) {
      throw const FormatException('Radius lokasi kantor tidak valid.');
    }
    if (isActive is! bool) {
      throw const FormatException('Status lokasi kantor tidak valid.');
    }

    return OfficeLocation(
      id: _string(json['id'], 'ID lokasi kantor tidak valid.'),
      name: _string(json['name'], 'Nama lokasi kantor tidak valid.'),
      address: address as String?,
      latitude: _number(
        json['latitude'],
        'Latitude lokasi kantor tidak valid.',
      ),
      longitude: _number(
        json['longitude'],
        'Longitude lokasi kantor tidak valid.',
      ),
      radiusMeters: radiusMeters,
      isActive: isActive,
    );
  }

  final String id;
  final String name;
  final String? address;
  final double latitude;
  final double longitude;
  final int radiusMeters;
  final bool isActive;
}

class LocationAssignmentRequirement {
  const LocationAssignmentRequirement({
    required this.id,
    required this.officeLocation,
    required this.effectiveFrom,
    required this.effectiveTo,
    required this.status,
  });

  factory LocationAssignmentRequirement.fromJson(Object? value) {
    final json = _object(value, 'Penugasan lokasi tidak valid.');
    return LocationAssignmentRequirement(
      id: _string(json['id'], 'ID penugasan lokasi tidak valid.'),
      officeLocation: OfficeLocation.fromJson(json['office_location']),
      effectiveFrom: _dateOnlyText(json['effective_from']),
      effectiveTo: _nullableDateOnlyText(json['effective_to']),
      status: _string(json['status'], 'Status penugasan lokasi tidak valid.'),
    );
  }

  final String id;
  final OfficeLocation officeLocation;
  final String effectiveFrom;
  final String? effectiveTo;
  final String status;
}

class LocationRequirement {
  const LocationRequirement({
    required this.assignment,
    required this.officeLocation,
  });

  factory LocationRequirement.fromJson(Object? value) {
    final json = _object(value, 'Kebutuhan lokasi tidak valid.');
    return LocationRequirement(
      assignment: LocationAssignmentRequirement.fromJson(json['assignment']),
      officeLocation: OfficeLocation.fromJson(json['office_location']),
    );
  }

  final LocationAssignmentRequirement assignment;
  final OfficeLocation officeLocation;
}

class AttendanceToday {
  const AttendanceToday({
    required this.attendanceDate,
    required this.schedule,
    required this.checkInAt,
    required this.checkOutAt,
    required this.state,
    required this.canCheckIn,
    required this.canCheckOut,
    required this.checkInLocation,
    required this.checkOutLocation,
  });

  factory AttendanceToday.fromJson(Object? value) {
    final json = _object(value, 'Data absensi tidak valid.');
    final canCheckIn = json['can_check_in'];
    final canCheckOut = json['can_check_out'];
    if (canCheckIn is! bool || canCheckOut is! bool) {
      throw const FormatException('Aksi absensi tidak valid.');
    }

    return AttendanceToday(
      attendanceDate: _attendanceDate(json['attendance_date']),
      schedule: WorkSchedule.fromJson(json['schedule']),
      checkInAt: _nullableDateTime(json['check_in_at']),
      checkOutAt: _nullableDateTime(json['check_out_at']),
      state: AttendanceState.fromJson(json['state']),
      canCheckIn: canCheckIn,
      canCheckOut: canCheckOut,
      checkInLocation: _nullableEvidence(json['check_in_location']),
      checkOutLocation: _nullableEvidence(json['check_out_location']),
    );
  }

  final DateTime attendanceDate;
  final WorkSchedule schedule;
  final DateTime? checkInAt;
  final DateTime? checkOutAt;
  final AttendanceState state;
  final bool canCheckIn;
  final bool canCheckOut;
  final AttendanceLocationEvidence? checkInLocation;
  final AttendanceLocationEvidence? checkOutLocation;
}

class AttendanceRecord {
  const AttendanceRecord({
    required this.id,
    required this.attendanceDate,
    required this.schedule,
    required this.checkInAt,
    required this.checkOutAt,
    required this.state,
    required this.checkInLocation,
    required this.checkOutLocation,
  });

  factory AttendanceRecord.fromJson(Object? value) {
    final json = _object(value, 'Record absensi tidak valid.');

    return AttendanceRecord(
      id: _string(json['id'], 'ID absensi tidak valid.'),
      attendanceDate: _attendanceDate(json['attendance_date']),
      schedule: WorkSchedule.fromJson(json['schedule']),
      checkInAt: _dateTime(json['check_in_at']),
      checkOutAt: _nullableDateTime(json['check_out_at']),
      state: AttendanceState.fromJson(json['state']),
      checkInLocation: _nullableEvidence(json['check_in_location']),
      checkOutLocation: _nullableEvidence(json['check_out_location']),
    );
  }

  final String id;
  final DateTime attendanceDate;
  final WorkSchedule schedule;
  final DateTime checkInAt;
  final DateTime? checkOutAt;
  final AttendanceState state;
  final AttendanceLocationEvidence? checkInLocation;
  final AttendanceLocationEvidence? checkOutLocation;
}

class AttendanceHistoryResponse {
  const AttendanceHistoryResponse({
    required this.items,
    required this.pagination,
  });

  factory AttendanceHistoryResponse.fromJson(Object? value) {
    final json = _object(value, 'Riwayat absensi tidak valid.');
    final rawItems = json['items'];
    if (rawItems is! List) {
      throw const FormatException('Daftar riwayat tidak valid.');
    }

    return AttendanceHistoryResponse(
      items: rawItems.map(AttendanceRecord.fromJson).toList(growable: false),
      pagination: AttendancePagination.fromJson(json),
    );
  }

  final List<AttendanceRecord> items;
  final AttendancePagination pagination;
}

class AttendancePagination {
  const AttendancePagination({
    required this.page,
    required this.pageSize,
    required this.totalItems,
    required this.totalPages,
  });

  factory AttendancePagination.fromJson(Object? value) {
    final json = _object(value, 'Pagination absensi tidak valid.');
    final page = json['page'];
    final pageSize = json['page_size'];
    final totalItems = json['total_items'];
    final totalPages = json['total_pages'];
    if (page is! int || page < 1) {
      throw const FormatException('Halaman riwayat tidak valid.');
    }
    if (pageSize is! int || pageSize < 1 || pageSize > 100) {
      throw const FormatException('Ukuran halaman riwayat tidak valid.');
    }
    if (totalItems is! int || totalItems < 0) {
      throw const FormatException('Total riwayat tidak valid.');
    }
    if (totalPages is! int || totalPages < 0) {
      throw const FormatException('Total halaman riwayat tidak valid.');
    }

    return AttendancePagination(
      page: page,
      pageSize: pageSize,
      totalItems: totalItems,
      totalPages: totalPages,
    );
  }

  final int page;
  final int pageSize;
  final int totalItems;
  final int totalPages;
}

Map<String, Object?> _object(Object? value, String message) {
  if (value is Map) {
    return Map<String, Object?>.from(value);
  }
  throw FormatException(message);
}

String _string(Object? value, String message) {
  if (value is String && value.trim().isNotEmpty) {
    return value;
  }
  throw FormatException(message);
}

double _number(Object? value, String message) {
  if (value is num && value.isFinite) {
    return value.toDouble();
  }
  throw FormatException(message);
}

bool _bool(Object? value, String message) {
  if (value is bool) {
    return value;
  }
  throw FormatException(message);
}

AttendanceLocationEvidence? _nullableEvidence(Object? value) {
  if (value == null) {
    return null;
  }
  return AttendanceLocationEvidence.fromJson(value);
}

DateTime _attendanceDate(Object? value) {
  final text = _dateOnlyText(value);
  return DateTime(
    int.parse(text.substring(0, 4)),
    int.parse(text.substring(5, 7)),
    int.parse(text.substring(8, 10)),
  );
}

String _dateOnlyText(Object? value) {
  final text = _string(value, 'Tanggal absensi tidak valid.');
  final date = DateTime.tryParse(text);
  if (date == null || text.length != 10) {
    throw const FormatException('Tanggal absensi tidak valid.');
  }
  return text;
}

String? _nullableDateOnlyText(Object? value) {
  if (value == null) {
    return null;
  }
  return _dateOnlyText(value);
}

DateTime _dateTime(Object? value) {
  final text = _string(value, 'Waktu absensi tidak valid.');
  final parsed = DateTime.tryParse(text);
  if (parsed == null) {
    throw const FormatException('Waktu absensi tidak valid.');
  }
  return parsed.toLocal();
}

DateTime? _nullableDateTime(Object? value) {
  if (value == null) {
    return null;
  }
  return _dateTime(value);
}
