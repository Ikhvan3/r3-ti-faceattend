import '../domain/face_attendance_grant.dart';
import '../domain/face_status.dart';
import '../domain/face_verification_result.dart';
import 'face_api_client.dart';

class FaceRepository {
  const FaceRepository({required FaceApi api}) : _api = api;

  final FaceApi _api;

  Future<FaceStatus> loadStatus() {
    return _api.getStatus();
  }

  Future<FaceStatus> enroll({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    return _api.enroll(
      embedding: embedding,
      embeddingModel: embeddingModel,
      embeddingVersion: embeddingVersion,
    );
  }

  Future<FaceVerificationResult> verify({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    return _api.verify(
      embedding: embedding,
      embeddingModel: embeddingModel,
      embeddingVersion: embeddingVersion,
    );
  }

  Future<FaceAttendanceGrant> verifyForAttendance({
    required String purpose,
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    return _api.verifyForAttendance(
      purpose: purpose,
      embedding: embedding,
      embeddingModel: embeddingModel,
      embeddingVersion: embeddingVersion,
    );
  }
}
