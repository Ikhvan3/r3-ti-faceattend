import '../domain/face_status.dart';
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

  Future<void> resetEnrollment() {
    return _api.resetEnrollment();
  }
}
