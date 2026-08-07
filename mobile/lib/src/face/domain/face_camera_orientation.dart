enum FaceCameraLens { front, back, external }

class FaceCameraOrientation {
  const FaceCameraOrientation({
    required this.lens,
    required this.sensorDegrees,
  });

  final FaceCameraLens lens;
  final int sensorDegrees;

  bool get isMirrored => lens == FaceCameraLens.front;

  // ML Kit headEulerAngleY is already expressed relative to the image being
  // processed: positive means the face turns to the camera/image right and
  // negative means it turns to the camera/image left. CameraPreview mirroring
  // is only a presentation concern and must not invert the detector value.
  double normalizeYaw(double yaw) => yaw;
}
