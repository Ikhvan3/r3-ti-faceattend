enum FaceCameraLens { front, back, external }

class FaceCameraOrientation {
  const FaceCameraOrientation({
    required this.lens,
    required this.sensorDegrees,
  });

  final FaceCameraLens lens;
  final int sensorDegrees;

  bool get isMirrored => lens == FaceCameraLens.front;

  double normalizeYaw(double yaw) {
    return isMirrored ? -yaw : yaw;
  }
}
