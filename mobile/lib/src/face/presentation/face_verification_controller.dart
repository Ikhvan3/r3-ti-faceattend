import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';

import '../data/face_detector_service.dart';
import '../data/face_embedding_service.dart';
import '../data/face_repository.dart';
import '../data/face_sample_collector.dart';
import '../domain/face_failure.dart';
import '../domain/face_model_config.dart';

enum FaceVerificationControllerStatus {
  initial,
  sampling,
  submitting,
  success,
  failure,
}

typedef FaceVerificationImageCapture = Future<String> Function();

class FaceVerificationController extends ChangeNotifier {
  FaceVerificationController({
    required FaceRepository repository,
    required FaceDetectorService detector,
    required FaceEmbeddingService embeddingService,
  }) : _repository = repository,
       _detector = detector,
       _embeddingService = embeddingService,
       _sampleCollector = FaceSampleCollector(
         detector: detector,
         embeddingService: embeddingService,
       );

  final FaceRepository _repository;
  final FaceDetectorService _detector;
  final FaceEmbeddingService _embeddingService;
  final FaceSampleCollector _sampleCollector;

  FaceVerificationControllerStatus _status =
      FaceVerificationControllerStatus.initial;
  String? _message;
  int _sampleCount = 0;
  bool _sessionExpired = false;
  bool _verified = false;
  bool _disposed = false;

  FaceVerificationControllerStatus get status => _status;
  String? get message => _message;
  int get sampleCount => _sampleCount;
  int get sampleTarget => FaceModelConfig.sampleTarget;
  bool get sessionExpired => _sessionExpired;
  bool get verified => _verified;
  bool get isBusy =>
      _status == FaceVerificationControllerStatus.sampling ||
      _status == FaceVerificationControllerStatus.submitting;

  Future<void> verifyFromCamera(FaceVerificationImageCapture capture) async {
    if (isBusy) {
      return;
    }
    final samples = <List<double>>[];
    _sampleCount = 0;
    _verified = false;
    _setState(
      status: FaceVerificationControllerStatus.sampling,
      message: 'Arahkan wajah ke kamera.',
    );

    while (samples.length < FaceModelConfig.sampleTarget && !_disposed) {
      String? imagePath;
      try {
        imagePath = await capture();
        final sample = await _sampleCollector.collectSample(imagePath);
        samples.add(sample);
        _sampleCount = samples.length;
        _setState(
          status: FaceVerificationControllerStatus.sampling,
          message: 'Sample $_sampleCount/${FaceModelConfig.sampleTarget}',
        );
        if (samples.length < FaceModelConfig.sampleTarget) {
          await Future<void>.delayed(FaceModelConfig.sampleInterval);
        }
      } on FaceFailure catch (error) {
        if (!_canRetrySample(error.kind)) {
          _applyFailure(error);
          return;
        }
        _setState(
          status: FaceVerificationControllerStatus.sampling,
          message: error.message,
        );
        await Future<void>.delayed(FaceModelConfig.sampleInterval);
      } finally {
        if (imagePath != null) {
          await _deleteTemporaryImage(imagePath);
        }
      }
    }

    if (_disposed) {
      return;
    }

    final aggregated = _sampleCollector.aggregateSamples(samples);
    _setState(
      status: FaceVerificationControllerStatus.submitting,
      message: 'Mengirim verifikasi wajah.',
    );
    try {
      final result = await _repository.verify(
        embedding: aggregated,
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      );
      _verified = result.verified;
      _setState(
        status: FaceVerificationControllerStatus.success,
        message: result.verified
            ? 'Wajah berhasil diverifikasi.'
            : 'Wajah tidak cocok dengan data yang terdaftar.',
      );
    } on FaceFailure catch (error) {
      _applyFailure(error);
    }
  }

  bool _canRetrySample(FaceFailureKind kind) {
    switch (kind) {
      case FaceFailureKind.noFace:
      case FaceFailureKind.multipleFaces:
      case FaceFailureKind.faceTooSmall:
      case FaceFailureKind.faceTooCloseToEdge:
      case FaceFailureKind.invalidPose:
        return true;
      case FaceFailureKind.cameraPermissionDenied:
      case FaceFailureKind.cameraUnavailable:
      case FaceFailureKind.corruptInput:
      case FaceFailureKind.invalidEmbedding:
      case FaceFailureKind.duplicateEnrollment:
      case FaceFailureKind.notEnrolled:
      case FaceFailureKind.verificationRejected:
      case FaceFailureKind.accountForbidden:
      case FaceFailureKind.sessionExpired:
      case FaceFailureKind.apiUnavailable:
      case FaceFailureKind.requestTimeout:
      case FaceFailureKind.malformedResponse:
      case FaceFailureKind.internalError:
        return false;
    }
  }

  void _applyFailure(FaceFailure error) {
    _sessionExpired = error.kind == FaceFailureKind.sessionExpired;
    _setState(
      status: FaceVerificationControllerStatus.failure,
      message: error.message,
    );
  }

  void _setState({
    required FaceVerificationControllerStatus status,
    String? message,
  }) {
    _status = status;
    _message = message;
    if (!_disposed) {
      notifyListeners();
    }
  }

  Future<void> _deleteTemporaryImage(String imagePath) async {
    try {
      await File(imagePath).delete();
    } catch (_) {
      // Camera plugin files are temporary; failure to delete is non-fatal.
    }
  }

  @override
  void dispose() {
    _disposed = true;
    unawaited(_detector.dispose());
    unawaited(_embeddingService.dispose());
    super.dispose();
  }
}
