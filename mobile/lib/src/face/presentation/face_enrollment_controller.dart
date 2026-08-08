import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';

import '../data/face_detector_service.dart';
import '../data/face_embedding_service.dart';
import '../data/face_repository.dart';
import '../data/face_sample_collector.dart';
import '../domain/face_failure.dart';
import '../domain/face_model_config.dart';
import '../domain/face_status.dart';

enum FaceControllerStatus {
  initial,
  loadingStatus,
  loaded,
  sampling,
  submitting,
  failure,
  success,
}

typedef FaceImageCapture = Future<String> Function();

class FaceEnrollmentController extends ChangeNotifier {
  FaceEnrollmentController({
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

  FaceControllerStatus _status = FaceControllerStatus.initial;
  FaceStatus? _faceStatus;
  String? _errorMessage;
  String? _qualityMessage;
  int _sampleCount = 0;
  bool _sessionExpired = false;
  bool _disposed = false;

  FaceControllerStatus get status => _status;
  FaceStatus? get faceStatus => _faceStatus;
  String? get errorMessage => _errorMessage;
  String? get qualityMessage => _qualityMessage;
  int get sampleCount => _sampleCount;
  int get sampleTarget => FaceModelConfig.sampleTarget;
  bool get sessionExpired => _sessionExpired;
  bool get isBusy =>
      _status == FaceControllerStatus.loadingStatus ||
      _status == FaceControllerStatus.sampling ||
      _status == FaceControllerStatus.submitting;

  Future<void> loadStatus() async {
    if (_status == FaceControllerStatus.loadingStatus) {
      return;
    }
    _setState(status: FaceControllerStatus.loadingStatus);
    try {
      _faceStatus = await _repository.loadStatus();
      _setState(status: FaceControllerStatus.loaded);
    } on FaceFailure catch (error) {
      _applyFailure(error);
    }
  }

  Future<void> enrollFromCamera(FaceImageCapture capture) async {
    if (isBusy) {
      return;
    }
    final samples = <List<double>>[];
    _sampleCount = 0;
    _setState(
      status: FaceControllerStatus.sampling,
      qualityMessage: 'Arahkan wajah ke kamera.',
    );

    while (samples.length < FaceModelConfig.sampleTarget && !_disposed) {
      String? imagePath;
      try {
        imagePath = await capture();
        final sample = await collectSample(imagePath);
        samples.add(sample);
        _sampleCount = samples.length;
        _setState(
          status: FaceControllerStatus.sampling,
          qualityMessage:
              'Sample $_sampleCount/${FaceModelConfig.sampleTarget}',
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
          status: FaceControllerStatus.sampling,
          qualityMessage: error.message,
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
    _setState(status: FaceControllerStatus.submitting);
    try {
      _faceStatus = await _repository.enroll(
        embedding: aggregated,
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      );
      _setState(
        status: FaceControllerStatus.success,
        qualityMessage: 'Enrollment wajah berhasil.',
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

  Future<List<double>> collectSample(String imagePath) async {
    return _sampleCollector.collectSample(imagePath);
  }

  void _applyFailure(FaceFailure error) {
    _sessionExpired = error.kind == FaceFailureKind.sessionExpired;
    _setState(
      status: FaceControllerStatus.failure,
      errorMessage: error.message,
      qualityMessage: error.message,
    );
  }

  void _setState({
    required FaceControllerStatus status,
    String? errorMessage,
    String? qualityMessage,
  }) {
    _status = status;
    _errorMessage = errorMessage;
    _qualityMessage = qualityMessage;
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
