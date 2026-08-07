import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/face/data/face_detector_service.dart';
import 'package:r3_ti_faceattend/src/face/data/face_embedding_service.dart';
import 'package:r3_ti_faceattend/src/face/data/face_repository.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_failure.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_model_config.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_verification_result.dart';
import 'package:r3_ti_faceattend/src/face/presentation/face_verification_controller.dart';

import 'face_controller_test.dart';
import 'face_test_fakes.dart';

void main() {
  test('permission denied and camera unavailable stop verification', () async {
    final denied = newFaceVerificationController();
    await denied.verifyFromCamera(
      () async => throw const FaceFailure(
        FaceFailureKind.cameraPermissionDenied,
        'denied',
      ),
    );

    expect(denied.status, FaceVerificationControllerStatus.failure);

    final unavailable = newFaceVerificationController();
    await unavailable.verifyFromCamera(
      () async => throw const FaceFailure(
        FaceFailureKind.cameraUnavailable,
        'unavailable',
      ),
    );

    expect(unavailable.status, FaceVerificationControllerStatus.failure);
  });

  test('no face, multiple faces, and invalid pose keep sampling', () async {
    final noFace = newFaceVerificationController(
      detector: FakeFaceDetector(faces: []),
    );
    final first = noFace.verifyFromCamera(() async => 'sample.jpg');
    await Future<void>.delayed(Duration.zero);
    expect(noFace.message, 'Wajah belum terdeteksi.');
    noFace.dispose();
    await first;

    final multiple = newFaceVerificationController(
      detector: FakeFaceDetector(faces: [validFace(), validFace()]),
    );
    final second = multiple.verifyFromCamera(() async => 'sample.jpg');
    await Future<void>.delayed(Duration.zero);
    expect(multiple.message, 'Pastikan hanya satu wajah di kamera.');
    multiple.dispose();
    await second;

    final pose = newFaceVerificationController(
      detector: FakeFaceDetector(faces: [validFace(headEulerAngleY: 40)]),
    );
    final third = pose.verifyFromCamera(() async => 'sample.jpg');
    await Future<void>.delayed(Duration.zero);
    expect(pose.message, 'Hadapkan wajah lurus ke kamera.');
    pose.dispose();
    await third;
  });

  test('valid sample calls embedding service and sends model version', () async {
    final api = FakeFaceApi(status: enrolledStatus);
    final embedding = FakeFaceEmbeddingService(
      embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 1),
    );
    final controller = newFaceVerificationController(
      api: api,
      embedding: embedding,
    );

    await controller.verifyFromCamera(() async => 'sample.jpg');

    expect(embedding.embedCount, FaceModelConfig.sampleTarget);
    expect(api.verifyCount, 1);
    expect(api.lastEmbedding, hasLength(FaceModelConfig.embeddingDimension));
    expect(api.lastEmbeddingModel, FaceModelConfig.identifier);
    expect(api.lastEmbeddingVersion, FaceModelConfig.version);
    expect(controller.verified, isTrue);
    expect(controller.message, 'Wajah berhasil diverifikasi.');
  });

  test('verified false UI state does not expire session', () async {
    final api = FakeFaceApi(
      status: enrolledStatus,
      verifyResult: const FaceVerificationResult(verified: false),
    );
    final controller = newFaceVerificationController(api: api);

    await controller.verifyFromCamera(() async => 'sample.jpg');

    expect(controller.verified, isFalse);
    expect(controller.sessionExpired, isFalse);
    expect(controller.message, 'Wajah tidak cocok dengan data yang terdaftar.');
  });

  test('API offline and invalid embedding fail safely', () async {
    final offline = newFaceVerificationController(
      api: FakeFaceApi(
        status: enrolledStatus,
        verifyError: const FaceFailure(FaceFailureKind.apiUnavailable, 'offline'),
      ),
    );
    await offline.verifyFromCamera(() async => 'sample.jpg');
    expect(offline.status, FaceVerificationControllerStatus.failure);

    final invalid = newFaceVerificationController(
      embedding: FakeFaceEmbeddingService(embedding: [1, 2, 3]),
    );
    await invalid.verifyFromCamera(() async => 'sample.jpg');
    expect(invalid.status, FaceVerificationControllerStatus.failure);
  });

  test('multiple submit is prevented and candidate is not persisted locally', () async {
    final api = FakeFaceApi(status: enrolledStatus);
    final controller = newFaceVerificationController(api: api);
    var captureCalls = 0;

    final first = controller.verifyFromCamera(() async {
      captureCalls += 1;
      await Future<void>.delayed(const Duration(milliseconds: 1));
      return 'sample.jpg';
    });
    await Future<void>.delayed(Duration.zero);
    await controller.verifyFromCamera(() async => 'ignored.jpg');
    await first;

    expect(captureCalls, FaceModelConfig.sampleTarget);
    expect(api.verifyCount, 1);
  });
}

FaceVerificationController newFaceVerificationController({
  FakeFaceApi? api,
  FaceDetectorService? detector,
  FaceEmbeddingService? embedding,
}) {
  return FaceVerificationController(
    repository: FaceRepository(api: api ?? FakeFaceApi(status: enrolledStatus)),
    detector: detector ?? FakeFaceDetector(faces: [validFace()]),
    embeddingService:
        embedding ??
        FakeFaceEmbeddingService(
          embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 1),
        ),
  );
}
