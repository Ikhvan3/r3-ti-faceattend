import '../domain/attendance_models.dart';
import 'attendance_api_client.dart';

class AttendanceRepository {
  const AttendanceRepository({required AttendanceApi api}) : _api = api;

  final AttendanceApi _api;

  Future<AttendanceToday> loadToday() {
    return _api.getTodayAttendance();
  }

  Future<AttendanceToday> checkIn(AttendanceLocationPayload location) {
    return _api.checkIn(location);
  }

  Future<AttendanceToday> checkOut(AttendanceLocationPayload location) {
    return _api.checkOut(location);
  }

  Future<LocationRequirement> loadLocationRequirement() {
    return _api.getLocationRequirement();
  }

  Future<AttendanceHistoryResponse> loadHistory({
    required int page,
    required int pageSize,
  }) {
    return _api.getAttendanceHistory(page: page, pageSize: pageSize);
  }
}
