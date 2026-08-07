import 'dart:io';

import 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart';
import 'package:image/image.dart' as img;

import '../domain/face_detection_result.dart';
import '../domain/face_failure.dart';

abstract class FaceDetectorService {
  Future<List<FaceDetectionResult>> detect(String imagePath);
  Future<void> dispose();
}

class MlKitFaceDetectorService implements FaceDetectorService {
  MlKitFaceDetectorService({FaceDetector? detector})
    : _detector =
          detector ??
          FaceDetector(
            options: FaceDetectorOptions(
              performanceMode: FaceDetectorMode.accurate,
              enableLandmarks: true,
              enableContours: false,
              enableClassification: false,
              enableTracking: false,
            ),
          );

  final FaceDetector _detector;

  @override
  Future<List<FaceDetectionResult>> detect(String imagePath) async {
    final bytes = await File(imagePath).readAsBytes();
    final image = img.decodeImage(bytes);
    if (image == null) {
      throw const FaceFailure(
        FaceFailureKind.corruptInput,
        'Gambar wajah tidak dapat dibaca.',
      );
    }
    final faces = await _detector.processImage(
      InputImage.fromFilePath(imagePath),
    );
    return faces
        .map(
          (face) => FaceDetectionResult(
            boundingBox: face.boundingBox,
            imageWidth: image.width,
            imageHeight: image.height,
            headEulerAngleY: face.headEulerAngleY,
            headEulerAngleZ: face.headEulerAngleZ,
          ),
        )
        .toList(growable: false);
  }

  @override
  Future<void> dispose() {
    return _detector.close();
  }
}
