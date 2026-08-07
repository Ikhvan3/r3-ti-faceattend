import 'dart:ui';

class FaceDetectionResult {
  const FaceDetectionResult({
    required this.boundingBox,
    required this.imageWidth,
    required this.imageHeight,
    this.headEulerAngleY,
    this.headEulerAngleZ,
  });

  final Rect boundingBox;
  final int imageWidth;
  final int imageHeight;
  final double? headEulerAngleY;
  final double? headEulerAngleZ;
}
