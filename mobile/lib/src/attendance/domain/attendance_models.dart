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

class AttendanceToday {
  const AttendanceToday({
    required this.attendanceDate,
    required this.schedule,
    required this.checkInAt,
    required this.checkOutAt,
    required this.state,
    required this.canCheckIn,
    required this.canCheckOut,
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
    );
  }

  final DateTime attendanceDate;
  final WorkSchedule schedule;
  final DateTime? checkInAt;
  final DateTime? checkOutAt;
  final AttendanceState state;
  final bool canCheckIn;
  final bool canCheckOut;
}

class AttendanceRecord {
  const AttendanceRecord({
    required this.id,
    required this.attendanceDate,
    required this.schedule,
    required this.checkInAt,
    required this.checkOutAt,
    required this.state,
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
    );
  }

  final String id;
  final DateTime attendanceDate;
  final WorkSchedule schedule;
  final DateTime checkInAt;
  final DateTime? checkOutAt;
  final AttendanceState state;
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

DateTime _attendanceDate(Object? value) {
  final text = _string(value, 'Tanggal absensi tidak valid.');
  final date = DateTime.tryParse(text);
  if (date == null || text.length != 10) {
    throw const FormatException('Tanggal absensi tidak valid.');
  }
  return DateTime(date.year, date.month, date.day);
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
