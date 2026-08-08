import 'dart:async';
import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/face/data/face_detector_service.dart';
import 'package:r3_ti_faceattend/src/face/data/face_embedding_service.dart';
import 'package:r3_ti_faceattend/src/face/data/face_repository.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_detection_result.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_embedding_math.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_failure.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_model_config.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_status.dart';
import 'package:r3_ti_faceattend/src/face/presentation/face_enrollment_controller.dart';

import 'face_test_fakes.dart';

void main() {
  test('loadStatus handles NOT_ENROLLED', () async {
    final api = FakeFaceApi(status: notEnrolledStatus);
    final controller = newFaceController(api: api);

    await controller.loadStatus();

    expect(controller.faceStatus?.status, FaceEnrollmentStatus.notEnrolled);
    expect(controller.status, FaceControllerStatus.loaded);
  });

  test('loadStatus handles ENROLLED', () async {
    final api = FakeFaceApi(status: enrolledStatus);
    final controller = newFaceController(api: api);

    await controller.loadStatus();

    expect(controller.faceStatus?.status, FaceEnrollmentStatus.enrolled);
    expect(controller.faceStatus?.enrolledAt, isNotNull);
  });

  test('API error does not mark session expired unless auth failure', () async {
    final api = FakeFaceApi(
      status: notEnrolledStatus,
      loadError: const FaceFailure(FaceFailureKind.apiUnavailable, 'offline'),
    );
    final controller = newFaceController(api: api);

    await controller.loadStatus();

    expect(controller.status, FaceControllerStatus.failure);
    expect(controller.sessionExpired, isFalse);
  });

  test('collectSample rejects zero and multiple faces', () async {
    final controller = newFaceController(detector: FakeFaceDetector(faces: []));
    await expectLater(
      controller.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.noFace,
        ),
      ),
    );

    final multiple = newFaceController(
      detector: FakeFaceDetector(faces: [validFace(), validFace()]),
    );
    await expectLater(
      multiple.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.multipleFaces,
        ),
      ),
    );
  });

  test('collectSample rejects small face and invalid pose', () async {
    final small = newFaceController(
      detector: FakeFaceDetector(
        faces: [validFace(box: const Rect.fromLTWH(100, 100, 40, 40))],
      ),
    );
    await expectLater(
      small.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.faceTooSmall,
        ),
      ),
    );

    final pose = newFaceController(
      detector: FakeFaceDetector(faces: [validFace(headEulerAngleY: 40)]),
    );
    await expectLater(
      pose.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.invalidPose,
        ),
      ),
    );
  });

  test('collectSample normalizes valid embedding', () async {
    final controller = newFaceController(
      detector: FakeFaceDetector(faces: [validFace()]),
      embedding: FakeFaceEmbeddingService(
        embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 2),
      ),
    );

    final sample = await controller.collectSample('unused');

    expect(sample, hasLength(FaceModelConfig.embeddingDimension));
    expect(
      sample.fold<double>(0, (sum, value) => sum + value * value),
      closeTo(1, 0.00001),
    );
  });

  test('collectSample rejects invalid embedding dimension and NaN', () async {
    final wrongDimension = newFaceController(
      detector: FakeFaceDetector(faces: [validFace()]),
      embedding: FakeFaceEmbeddingService(embedding: [1, 2, 3]),
    );
    await expectLater(
      wrongDimension.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.invalidEmbedding,
        ),
      ),
    );

    final nan = newFaceController(
      detector: FakeFaceDetector(faces: [validFace()]),
      embedding: FakeFaceEmbeddingService(
        embedding: List<double>.filled(
          FaceModelConfig.embeddingDimension,
          double.nan,
        ),
      ),
    );
    await expectLater(
      nan.collectSample('unused'),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.invalidEmbedding,
        ),
      ),
    );
  });

  test('averaging normalizes final embedding', () {
    final a = l2NormalizeEmbedding(
      List<double>.filled(FaceModelConfig.embeddingDimension, 1),
    );
    final b = l2NormalizeEmbedding(
      List<double>.filled(FaceModelConfig.embeddingDimension, 2),
    );

    final averaged = averageEmbeddings([
      a,
      b,
    ], dimension: FaceModelConfig.embeddingDimension);

    expect(
      averaged.fold<double>(0, (sum, value) => sum + value * value),
      closeTo(1, 0.00001),
    );
  });

  test(
    'enrollFromCamera submits aggregate and prevents duplicate submit',
    () async {
      final api = FakeFaceApi(status: notEnrolledStatus);
      final controller = newFaceController(
        api: api,
        detector: FakeFaceDetector(faces: [validFace()]),
        embedding: FakeFaceEmbeddingService(
          embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 1),
        ),
      );

      var captureCount = 0;
      await controller.enrollFromCamera(
        () async => 'sample-${captureCount++}.jpg',
      );

      expect(api.enrollCount, 1);
      expect(api.lastEmbedding, hasLength(FaceModelConfig.embeddingDimension));
      expect(controller.status, FaceControllerStatus.success);
    },
  );

  test(
    'camera permission denied and camera unavailable stop enrollment',
    () async {
      final denied = newFaceController();
      await denied.enrollFromCamera(
        () async => throw const FaceFailure(
          FaceFailureKind.cameraPermissionDenied,
          'camera denied',
        ),
      );
      expect(denied.status, FaceControllerStatus.failure);
      expect(denied.errorMessage, 'camera denied');

      final unavailable = newFaceController();
      await unavailable.enrollFromCamera(
        () async => throw const FaceFailure(
          FaceFailureKind.cameraUnavailable,
          'camera unavailable',
        ),
      );
      expect(unavailable.status, FaceControllerStatus.failure);
      expect(unavailable.errorMessage, 'camera unavailable');
    },
  );

  test('multiple submit is ignored while sampling', () async {
    final api = FakeFaceApi(status: notEnrolledStatus);
    final controller = newFaceController(
      api: api,
      detector: FakeFaceDetector(faces: [validFace()]),
    );
    final gate = Completer<void>();
    var captureCalls = 0;

    final first = controller.enrollFromCamera(() async {
      captureCalls += 1;
      await gate.future;
      return 'sample.jpg';
    });
    await Future<void>.delayed(Duration.zero);
    await controller.enrollFromCamera(() async => 'ignored.jpg');
    gate.complete();
    await first;

    expect(captureCalls, FaceModelConfig.sampleTarget);
    expect(api.enrollCount, 1);
  });

  test('enrollFromCamera maps duplicate and offline errors', () async {
    final duplicate = FakeFaceApi(
      status: notEnrolledStatus,
      enrollError: const FaceFailure(
        FaceFailureKind.duplicateEnrollment,
        'duplicate',
      ),
    );
    final controller = newFaceController(
      api: duplicate,
      detector: FakeFaceDetector(faces: [validFace()]),
    );

    await controller.enrollFromCamera(() async => 'sample.jpg');

    expect(controller.status, FaceControllerStatus.failure);
    expect(controller.errorMessage, 'duplicate');
  });
}

FaceEnrollmentController newFaceController({
  FakeFaceApi? api,
  FaceDetectorService? detector,
  FaceEmbeddingService? embedding,
}) {
  return FaceEnrollmentController(
    repository: FaceRepository(
      api: api ?? FakeFaceApi(status: notEnrolledStatus),
    ),
    detector: detector ?? FakeFaceDetector(faces: [validFace()]),
    embeddingService:
        embedding ??
        FakeFaceEmbeddingService(
          embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 1),
        ),
  );
}

FaceDetectionResult validFace({
  Rect box = const Rect.fromLTWH(80, 80, 220, 220),
  double? headEulerAngleY,
  double? headEulerAngleZ,
}) {
  return FaceDetectionResult(
    boundingBox: box,
    imageWidth: 400,
    imageHeight: 400,
    headEulerAngleY: headEulerAngleY,
    headEulerAngleZ: headEulerAngleZ,
  );
}
