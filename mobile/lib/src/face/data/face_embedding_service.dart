import 'dart:io';
import 'dart:math';

import 'package:image/image.dart' as img;
import 'package:tflite_flutter/tflite_flutter.dart';

import '../domain/face_detection_result.dart';
import '../domain/face_failure.dart';
import '../domain/face_model_config.dart';

abstract class FaceEmbeddingService {
  Future<List<double>> embed({
    required String imagePath,
    required FaceDetectionResult face,
  });
  Future<void> dispose();
}

class TfliteFaceEmbeddingService implements FaceEmbeddingService {
  TfliteFaceEmbeddingService({
    Future<Interpreter> Function()? interpreterLoader,
  }) : _interpreterLoader =
           interpreterLoader ??
           (() => Interpreter.fromAsset(FaceModelConfig.assetPath));

  final Future<Interpreter> Function() _interpreterLoader;
  Future<Interpreter>? _interpreter;

  @override
  Future<List<double>> embed({
    required String imagePath,
    required FaceDetectionResult face,
  }) async {
    final bytes = await File(imagePath).readAsBytes();
    final decoded = img.decodeImage(bytes);
    if (decoded == null) {
      throw const FaceFailure(
        FaceFailureKind.corruptInput,
        'Gambar wajah tidak dapat dibaca.',
      );
    }

    final input = _preprocess(img.bakeOrientation(decoded), face);
    final output = List<List<double>>.generate(
      1,
      (_) => List<double>.filled(FaceModelConfig.embeddingDimension, 0),
    );
    final interpreter = await (_interpreter ??= _interpreterLoader());
    interpreter.run(input, output);
    final embedding = output.first;
    if (embedding.length != FaceModelConfig.embeddingDimension ||
        embedding.any((value) => !value.isFinite)) {
      throw const FaceFailure(
        FaceFailureKind.invalidEmbedding,
        'Embedding wajah tidak valid.',
      );
    }
    return embedding;
  }

  List<List<List<List<double>>>> _preprocess(
    img.Image image,
    FaceDetectionResult face,
  ) {
    final crop = _cropFace(image, face);
    final resized = img.copyResize(
      crop,
      width: FaceModelConfig.inputWidth,
      height: FaceModelConfig.inputHeight,
      interpolation: img.Interpolation.linear,
    );
    return [
      List.generate(FaceModelConfig.inputHeight, (y) {
        return List.generate(FaceModelConfig.inputWidth, (x) {
          final pixel = resized.getPixel(x, y);
          return <double>[
            (pixel.r.toDouble() - 127.5) / 127.5,
            (pixel.g.toDouble() - 127.5) / 127.5,
            (pixel.b.toDouble() - 127.5) / 127.5,
          ];
        }, growable: false);
      }, growable: false),
    ];
  }

  img.Image _cropFace(img.Image image, FaceDetectionResult face) {
    final box = face.boundingBox;
    final marginX = box.width * FaceModelConfig.cropMarginRatio;
    final marginY = box.height * FaceModelConfig.cropMarginRatio;
    final left = max(0, (box.left - marginX).floor());
    final top = max(0, (box.top - marginY).floor());
    final right = min(image.width, (box.right + marginX).ceil());
    final bottom = min(image.height, (box.bottom + marginY).ceil());
    final width = right - left;
    final height = bottom - top;
    if (width <= 0 || height <= 0) {
      throw const FaceFailure(
        FaceFailureKind.corruptInput,
        'Crop wajah tidak valid.',
      );
    }
    return img.copyCrop(image, x: left, y: top, width: width, height: height);
  }

  @override
  Future<void> dispose() async {
    final interpreter = _interpreter;
    if (interpreter == null) {
      return;
    }
    (await interpreter).close();
  }
}
