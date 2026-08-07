import 'dart:ui';

import 'face_detection_result.dart';

class FaceLivenessObservation {
  const FaceLivenessObservation({
    required this.boundingBox,
    required this.imageWidth,
    required this.imageHeight,
    this.trackingId,
    this.headEulerAngleY,
    this.headEulerAngleZ,
    this.leftEyeOpenProbability,
    this.rightEyeOpenProbability,
  });

  factory FaceLivenessObservation.fromDetection(FaceDetectionResult result) {
    return FaceLivenessObservation(
      boundingBox: result.boundingBox,
      imageWidth: result.imageWidth,
      imageHeight: result.imageHeight,
      trackingId: result.trackingId,
      headEulerAngleY: result.headEulerAngleY,
      headEulerAngleZ: result.headEulerAngleZ,
      leftEyeOpenProbability: result.leftEyeOpenProbability,
      rightEyeOpenProbability: result.rightEyeOpenProbability,
    );
  }

  final Rect boundingBox;
  final int imageWidth;
  final int imageHeight;
  final int? trackingId;
  final double? headEulerAngleY;
  final double? headEulerAngleZ;
  final double? leftEyeOpenProbability;
  final double? rightEyeOpenProbability;
}
