import 'dart:io';

import 'package:camera/camera.dart';
import 'package:flutter/services.dart';
import 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart';
import 'package:image/image.dart' as img;

import '../domain/face_camera_orientation.dart';
import '../domain/face_detection_result.dart';
import '../domain/face_failure.dart';

abstract class FaceDetectorService {
  Future<List<FaceDetectionResult>> detect(String imagePath);
  Future<List<FaceDetectionResult>> detectCameraImage({
    required CameraImage image,
    required CameraDescription camera,
    required DeviceOrientation deviceOrientation,
  });
  Future<void> dispose();
}

class MlKitFaceDetectorService implements FaceDetectorService {
  MlKitFaceDetectorService({FaceDetector? detector})
    : _detector = detector ?? FaceDetector(options: _recognitionOptions);

  MlKitFaceDetectorService.liveness({FaceDetector? detector})
    : _detector = detector ?? FaceDetector(options: _livenessOptions);

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
        .map((face) => _mapFace(face, image.width, image.height))
        .toList(growable: false);
  }

  @override
  Future<List<FaceDetectionResult>> detectCameraImage({
    required CameraImage image,
    required CameraDescription camera,
    required DeviceOrientation deviceOrientation,
  }) async {
    final rotation = _imageRotation(camera, deviceOrientation);
    if (rotation == null || image.planes.isEmpty) {
      return const [];
    }

    final inputImage = _buildInputImage(image: image, rotation: rotation);
    if (inputImage == null) {
      return const [];
    }

    final faces = await _detector.processImage(inputImage);

    final isRotated =
        rotation == InputImageRotation.rotation90deg ||
        rotation == InputImageRotation.rotation270deg;
    final effectiveWidth = isRotated ? image.height : image.width;
    final effectiveHeight = isRotated ? image.width : image.height;

    return faces
        .map((face) => _mapFace(face, effectiveWidth, effectiveHeight))
        .toList(growable: false);
  }

  InputImage? _buildInputImage({
    required CameraImage image,
    required InputImageRotation rotation,
  }) {
    final Uint8List bytes;
    final InputImageFormat format;

    if (Platform.isAndroid) {
      bytes = _yuv420ToNv21(image);
      format = InputImageFormat.nv21;
    } else {
      final resolvedFormat = InputImageFormatValue.fromRawValue(
        image.format.raw,
      );
      if (resolvedFormat == null) return null;
      final writeBuffer = WriteBuffer();
      for (final plane in image.planes) {
        writeBuffer.putUint8List(plane.bytes);
      }
      bytes = writeBuffer.done().buffer.asUint8List();
      format = resolvedFormat;
    }

    return InputImage.fromBytes(
      bytes: bytes,
      metadata: InputImageMetadata(
        size: Size(image.width.toDouble(), image.height.toDouble()),
        rotation: rotation,
        format: format,
        bytesPerRow: image.width,
      ),
    );
  }

  FaceDetectionResult _mapFace(Face face, int imageWidth, int imageHeight) {
    return FaceDetectionResult(
      boundingBox: face.boundingBox,
      imageWidth: imageWidth,
      imageHeight: imageHeight,
      trackingId: face.trackingId,
      leftEyeOpenProbability: face.leftEyeOpenProbability,
      rightEyeOpenProbability: face.rightEyeOpenProbability,
      headEulerAngleY: face.headEulerAngleY,
      headEulerAngleZ: face.headEulerAngleZ,
    );
  }

  Uint8List _yuv420ToNv21(CameraImage image) {
    final width = image.width;
    final height = image.height;
    final yPlane = image.planes[0];
    final uPlane = image.planes[1];
    final vPlane = image.planes[2];

    final ySize = width * height;
    final uvSize = width * height ~/ 2;
    final nv21 = Uint8List(ySize + uvSize);

    // Copy Y plane, stripping row padding if bytesPerRow > width.
    var offset = 0;
    for (var row = 0; row < height; row++) {
      final rowStart = row * yPlane.bytesPerRow;
      nv21.setRange(offset, offset + width, yPlane.bytes, rowStart);
      offset += width;
    }

    // Interleave V and U planes into NV21 (VU order), respecting pixel/row stride.
    final uvRowStride = uPlane.bytesPerRow;
    final uvPixelStride = uPlane.bytesPerPixel ?? 2;
    final chromaHeight = height ~/ 2;
    final chromaWidth = width ~/ 2;

    for (var row = 0; row < chromaHeight; row++) {
      var uRowStart = row * uvRowStride;
      var vRowStart = row * vPlane.bytesPerRow;
      for (var col = 0; col < chromaWidth; col++) {
        final uIndex = uRowStart + col * uvPixelStride;
        final vIndex = vRowStart + col * uvPixelStride;
        nv21[offset++] = vPlane.bytes[vIndex];
        nv21[offset++] = uPlane.bytes[uIndex];
      }
    }

    return nv21;
  }

  InputImageRotation? _imageRotation(
    CameraDescription camera,
    DeviceOrientation deviceOrientation,
  ) {
    final rotationDegrees = _rotationDegrees[deviceOrientation];
    if (rotationDegrees == null) {
      return null;
    }
    final sensorOrientation = camera.sensorOrientation;
    final rotationCompensation =
        camera.lensDirection == CameraLensDirection.front
        ? (sensorOrientation + rotationDegrees) % 360
        : (sensorOrientation - rotationDegrees + 360) % 360;
    return InputImageRotationValue.fromRawValue(rotationCompensation);
  }

  @override
  Future<void> dispose() {
    return _detector.close();
  }
}

const _rotationDegrees = <DeviceOrientation, int>{
  DeviceOrientation.portraitUp: 0,
  DeviceOrientation.landscapeLeft: 90,
  DeviceOrientation.portraitDown: 180,
  DeviceOrientation.landscapeRight: 270,
};

final _recognitionOptions = FaceDetectorOptions(
  performanceMode: FaceDetectorMode.accurate,
  enableLandmarks: false,
  enableContours: false,
  enableClassification: true,
  enableTracking: true,
);

final _livenessOptions = FaceDetectorOptions(
  performanceMode: FaceDetectorMode.fast,
  enableLandmarks: false,
  enableContours: false,
  enableClassification: true,
  enableTracking: true,
);

FaceCameraOrientation faceCameraOrientationFromDescription(
  CameraDescription camera,
) {
  final lens = switch (camera.lensDirection) {
    CameraLensDirection.front => FaceCameraLens.front,
    CameraLensDirection.back => FaceCameraLens.back,
    CameraLensDirection.external => FaceCameraLens.external,
  };
  return FaceCameraOrientation(
    lens: lens,
    sensorDegrees: camera.sensorOrientation,
  );
}
