import 'dart:math';

import '../domain/face_detection_result.dart';
import '../domain/face_embedding_math.dart';
import '../domain/face_failure.dart';
import '../domain/face_model_config.dart';
import 'face_detector_service.dart';
import 'face_embedding_service.dart';

class FaceSampleCollector {
  const FaceSampleCollector({
    required FaceDetectorService detector,
    required FaceEmbeddingService embeddingService,
  }) : _detector = detector,
       _embeddingService = embeddingService;

  final FaceDetectorService _detector;
  final FaceEmbeddingService _embeddingService;

  Future<List<double>> collectSample(String imagePath) async {
    final faces = await _detector.detect(imagePath);
    final face = _validateFaces(faces);
    final embedding = await _embeddingService.embed(
      imagePath: imagePath,
      face: face,
    );
    if (embedding.length != FaceModelConfig.embeddingDimension) {
      throw const FaceFailure(
        FaceFailureKind.invalidEmbedding,
        'Dimensi embedding wajah tidak valid.',
      );
    }
    return l2NormalizeEmbedding(embedding);
  }

  List<double> aggregateSamples(List<List<double>> samples) {
    return averageEmbeddings(
      samples,
      dimension: FaceModelConfig.embeddingDimension,
    );
  }

  FaceDetectionResult _validateFaces(List<FaceDetectionResult> faces) {
    return validateFaceSample(faces);
  }
}

FaceDetectionResult validateFaceSample(List<FaceDetectionResult> faces) {
  if (faces.isEmpty) {
    throw const FaceFailure(FaceFailureKind.noFace, 'Wajah belum terdeteksi.');
  }
  if (faces.length > 1) {
    throw const FaceFailure(
      FaceFailureKind.multipleFaces,
      'Pastikan hanya satu wajah di kamera.',
    );
  }
  final face = faces.single;
  final imageShortSide = min(face.imageWidth, face.imageHeight);
  final boxShortSide = min(face.boundingBox.width, face.boundingBox.height);
  if (boxShortSide / imageShortSide < FaceModelConfig.minFaceBoxRatio) {
    throw const FaceFailure(
      FaceFailureKind.faceTooSmall,
      'Dekatkan wajah ke kamera.',
    );
  }
  final marginX = face.imageWidth * FaceModelConfig.edgeMarginRatio;
  final marginY = face.imageHeight * FaceModelConfig.edgeMarginRatio;
  if (face.boundingBox.left < marginX ||
      face.boundingBox.top < marginY ||
      face.boundingBox.right > face.imageWidth - marginX ||
      face.boundingBox.bottom > face.imageHeight - marginY) {
    throw const FaceFailure(
      FaceFailureKind.faceTooCloseToEdge,
      'Posisikan wajah di tengah frame.',
    );
  }
  final yaw = face.headEulerAngleY?.abs() ?? 0;
  final roll = face.headEulerAngleZ?.abs() ?? 0;
  if (yaw > FaceModelConfig.maxHeadEulerY ||
      roll > FaceModelConfig.maxHeadEulerZ) {
    throw const FaceFailure(
      FaceFailureKind.invalidPose,
      'Hadapkan wajah lurus ke kamera.',
    );
  }
  return face;
}
