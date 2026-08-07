import 'dart:ui';

class FaceDetectionResult {
  const FaceDetectionResult({
    required this.boundingBox,
    required this.imageWidth,
    required this.imageHeight,
    this.trackingId,
    this.leftEyeOpenProbability,
    this.rightEyeOpenProbability,
    this.headEulerAngleY,
    this.headEulerAngleZ,
  });

  final Rect boundingBox;
  final int imageWidth;
  final int imageHeight;
  final int? trackingId;
  final double? leftEyeOpenProbability;
  final double? rightEyeOpenProbability;
  final double? headEulerAngleY;
  final double? headEulerAngleZ;
}
