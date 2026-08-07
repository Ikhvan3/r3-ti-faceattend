import 'package:r3_ti_faceattend/src/face/data/face_api_client.dart';
import 'package:r3_ti_faceattend/src/face/data/face_detector_service.dart';
import 'package:r3_ti_faceattend/src/face/data/face_embedding_service.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_detection_result.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_failure.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_model_config.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_status.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_verification_result.dart';

final notEnrolledStatus = FaceStatus.fromJson(<String, Object?>{
  'enrolled': false,
  'face_status': 'NOT_ENROLLED',
});

final enrolledStatus = FaceStatus.fromJson(<String, Object?>{
  'enrolled': true,
  'face_status': 'ENROLLED',
  'embedding_model': FaceModelConfig.identifier,
  'embedding_version': FaceModelConfig.version,
  'enrolled_at': '2026-08-07T01:00:00Z',
});

class FakeFaceApi implements FaceApi {
  FakeFaceApi({
    required this.status,
    this.loadError,
    this.enrollError,
    this.verifyError,
    this.verifyResult = const FaceVerificationResult(verified: true),
    this.resetError,
  });

  FaceStatus status;
  FaceFailure? loadError;
  FaceFailure? enrollError;
  FaceFailure? verifyError;
  FaceVerificationResult verifyResult;
  FaceFailure? resetError;
  int enrollCount = 0;
  int verifyCount = 0;
  int resetCount = 0;
  List<double>? lastEmbedding;
  String? lastEmbeddingModel;
  String? lastEmbeddingVersion;

  @override
  Future<FaceStatus> getStatus() async {
    final error = loadError;
    if (error != null) {
      throw error;
    }
    return status;
  }

  @override
  Future<FaceStatus> enroll({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) async {
    enrollCount += 1;
    lastEmbedding = embedding;
    lastEmbeddingModel = embeddingModel;
    lastEmbeddingVersion = embeddingVersion;
    final error = enrollError;
    if (error != null) {
      throw error;
    }
    status = enrolledStatus;
    return status;
  }

  @override
  Future<FaceVerificationResult> verify({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) async {
    verifyCount += 1;
    lastEmbedding = embedding;
    lastEmbeddingModel = embeddingModel;
    lastEmbeddingVersion = embeddingVersion;
    final error = verifyError;
    if (error != null) {
      throw error;
    }
    return verifyResult;
  }

  @override
  Future<void> resetEnrollment() async {
    resetCount += 1;
    final error = resetError;
    if (error != null) {
      throw error;
    }
    status = notEnrolledStatus;
  }
}

class FakeFaceDetector implements FaceDetectorService {
  FakeFaceDetector({required this.faces, this.error});

  List<FaceDetectionResult> faces;
  FaceFailure? error;
  bool disposed = false;

  @override
  Future<List<FaceDetectionResult>> detect(String imagePath) async {
    final failure = error;
    if (failure != null) {
      throw failure;
    }
    return faces;
  }

  @override
  Future<void> dispose() async {
    disposed = true;
  }
}

class FakeFaceEmbeddingService implements FaceEmbeddingService {
  FakeFaceEmbeddingService({required this.embedding, this.error});

  List<double> embedding;
  FaceFailure? error;
  bool disposed = false;
  int embedCount = 0;

  @override
  Future<List<double>> embed({
    required String imagePath,
    required FaceDetectionResult face,
  }) async {
    embedCount += 1;
    final failure = error;
    if (failure != null) {
      throw failure;
    }
    return embedding;
  }

  @override
  Future<void> dispose() async {
    disposed = true;
  }
}
