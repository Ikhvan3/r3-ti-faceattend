import 'dart:async';
import 'dart:io';
import 'dart:math';

import 'package:flutter/foundation.dart';

import '../data/face_detector_service.dart';
import '../data/face_embedding_service.dart';
import '../data/face_repository.dart';
import '../data/face_sample_collector.dart';
import '../domain/face_camera_orientation.dart';
import '../domain/face_detection_result.dart';
import '../domain/face_failure.dart';
import '../domain/face_liveness_models.dart';
import '../domain/face_liveness_observation.dart';
import '../domain/face_model_config.dart';

typedef LivenessImageCapture = Future<String> Function();
typedef LivenessNow = DateTime Function();

enum _BlinkPhase { waitingOpen, waitingClosed, waitingReopen }

class FaceLivenessController extends ChangeNotifier {
  FaceLivenessController({
    required FaceRepository repository,
    required FaceDetectorService detector,
    required FaceEmbeddingService embeddingService,
    FaceCameraOrientation cameraOrientation = const FaceCameraOrientation(
      lens: FaceCameraLens.front,
      sensorDegrees: 0,
    ),
    LivenessConfig config = const LivenessConfig(),
    Random? random,
    LivenessNow? now,
  }) : _repository = repository,
       _detector = detector,
       _embeddingService = embeddingService,
       _cameraOrientation = cameraOrientation,
       _config = config,
       _random = random ?? Random.secure(),
       _now = now ?? DateTime.now,
       _sampleCollector = FaceSampleCollector(
         detector: detector,
         embeddingService: embeddingService,
       );

  final FaceRepository _repository;
  final FaceDetectorService _detector;
  final FaceEmbeddingService _embeddingService;
  final FaceSampleCollector _sampleCollector;
  final FaceCameraOrientation _cameraOrientation;
  final LivenessConfig _config;
  final Random _random;
  final LivenessNow _now;

  LivenessResult _state = const LivenessResult(
    status: LivenessResultStatus.initializing,
    message: 'Posisikan wajah di tengah.',
  );
  DateTime? _challengeStartedAt;
  DateTime? _stepStartedAt;
  DateTime? _lastValidFaceAt;
  DateTime? _stabilizingSince;
  DateTime? _multipleFaceSince;
  DateTime? _trackingMismatchSince;
  DateTime? _actionStableSince;
  DateTime? _lastActionMatchAt;
  int? _trackingId;
  _BlinkPhase _blinkPhase = _BlinkPhase.waitingOpen;
  bool _disposed = false;
  bool _isSubmitting = false;

  LivenessResult get state => _state;
  bool get isBusy =>
      _state.status == LivenessResultStatus.stabilizing ||
      _state.status == LivenessResultStatus.executingAction ||
      _state.status == LivenessResultStatus.returningCenter ||
      _state.status == LivenessResultStatus.completed ||
      _state.status == LivenessResultStatus.verifying;
  bool get canRetry =>
      _state.status == LivenessResultStatus.failure ||
      _state.status == LivenessResultStatus.success ||
      _state.status == LivenessResultStatus.initializing;

  void start() {
    if (_state.status == LivenessResultStatus.verifying || _isSubmitting) {
      return;
    }
    _resetSession(
      status: LivenessResultStatus.waitingForFace,
      message: 'Posisikan wajah di tengah.',
      stepState: LivenessStepState.waitingForFace,
    );
  }

  void retry() {
    start();
  }

  void processFaces(List<FaceDetectionResult> faces) {
    processObservations(
      faces.map(FaceLivenessObservation.fromDetection).toList(growable: false),
    );
  }

  void processObservations(List<FaceLivenessObservation> observations) {
    if (_disposed ||
        _state.status == LivenessResultStatus.verifying ||
        _state.status == LivenessResultStatus.success ||
        _state.status == LivenessResultStatus.completed ||
        _state.status == LivenessResultStatus.failure) {
      return;
    }

    final now = _now();
    if (observations.isEmpty) {
      _handleFaceLost(now, 'Posisikan wajah di tengah.');
      return;
    }
    if (observations.length > 1) {
      _handleMultipleFaces(now);
      return;
    }
    _multipleFaceSince = null;

    final observation = observations.single;
    if (!_hasObservableQuality(observation)) {
      _handleFaceLost(now, 'Posisikan wajah di tengah.');
      return;
    }

    switch (_state.status) {
      case LivenessResultStatus.initializing:
      case LivenessResultStatus.waitingForFace:
      case LivenessResultStatus.stabilizing:
        _handleStabilization(now, observation);
      case LivenessResultStatus.executingAction:
      case LivenessResultStatus.returningCenter:
        _handleActiveChallenge(now, observation);
      case LivenessResultStatus.completed:
      case LivenessResultStatus.verifying:
      case LivenessResultStatus.success:
      case LivenessResultStatus.failure:
        return;
    }
  }

  void failFromDetector(String message) {
    if (_disposed || _state.status == LivenessResultStatus.failure) {
      return;
    }
    _fail(message);
  }

  Future<void> verifyAfterLiveness(LivenessImageCapture capture) async {
    if (_state.status != LivenessResultStatus.completed || _isSubmitting) {
      return;
    }
    _isSubmitting = true;
    final samples = <List<double>>[];
    _setState(
      _state.copyWith(
        status: LivenessResultStatus.verifying,
        message: 'Memverifikasi wajah...',
      ),
    );

    while (samples.length < FaceModelConfig.sampleTarget && !_disposed) {
      String? imagePath;
      try {
        imagePath = await capture();
        samples.add(await _sampleCollector.collectSample(imagePath));
        if (samples.length < FaceModelConfig.sampleTarget) {
          await Future<void>.delayed(_config.sampleInterval);
        }
      } on FaceFailure catch (error) {
        if (!_canRetrySample(error.kind)) {
          _applyFailure(error);
          return;
        }
        _setState(_state.copyWith(message: error.message));
        await Future<void>.delayed(_config.sampleInterval);
      } finally {
        if (imagePath != null) {
          await _deleteTemporaryImage(imagePath);
        }
      }
    }

    if (_disposed) {
      return;
    }

    try {
      final result = await _repository.verify(
        embedding: _sampleCollector.aggregateSamples(samples),
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      );
      _setState(
        _state.copyWith(
          status: LivenessResultStatus.success,
          message: result.verified
              ? 'Verifikasi wajah berhasil.'
              : 'Wajah tidak sesuai dengan data yang terdaftar.',
          verified: result.verified,
        ),
      );
    } on FaceFailure catch (error) {
      _applyFailure(error);
    } finally {
      _isSubmitting = false;
    }
  }

  void _handleStabilization(DateTime now, FaceLivenessObservation observation) {
    _lastValidFaceAt = now;
    if (!_isPoseCenter(observation)) {
      _stabilizingSince = null;
      _setState(
        _state.copyWith(
          status: LivenessResultStatus.waitingForFace,
          message: 'Hadapkan wajah lurus ke kamera.',
          stepState: LivenessStepState.waitingForFace,
        ),
      );
      _debugObservation(now, observation);
      return;
    }

    _stabilizingSince ??= now;
    final stableDuration = now.difference(_stabilizingSince!);
    if (stableDuration < _config.stabilizationDuration) {
      _setState(
        _state.copyWith(
          status: LivenessResultStatus.stabilizing,
          message: 'Tahan wajah di tengah.',
          stepState: LivenessStepState.stabilizing,
        ),
      );
      _debugObservation(now, observation);
      return;
    }

    _beginChallenge(now, observation);
  }

  void _handleActiveChallenge(
    DateTime now,
    FaceLivenessObservation observation,
  ) {
    if (_hasTimedOut(now)) {
      _fail('Verifikasi keaktifan gagal. Silakan coba lagi.');
      return;
    }
    if (!_hasSameTracking(now, observation)) {
      return;
    }

    _lastValidFaceAt = now;
    final action = _state.status == LivenessResultStatus.returningCenter
        ? LivenessAction.returnCenter
        : _state.currentAction;
    if (action == null) {
      return;
    }

    final matched = _matchesAction(action, observation, now);
    _debugObservation(now, observation);
    if (matched) {
      _completeStep(now, action);
    }
  }

  bool _hasObservableQuality(FaceLivenessObservation observation) {
    final imageShortSide = min(observation.imageWidth, observation.imageHeight);
    final boxShortSide = min(
      observation.boundingBox.width,
      observation.boundingBox.height,
    );
    if (boxShortSide / imageShortSide < _config.minFaceBoxRatio) {
      return false;
    }
    if (!_isCentered(observation)) {
      return false;
    }
    return (observation.headEulerAngleZ?.abs() ?? 0) <= _config.maxRollDegrees;
  }

  bool _isCentered(FaceLivenessObservation observation) {
    final marginX = observation.imageWidth * _config.edgeMarginRatio;
    final marginY = observation.imageHeight * _config.edgeMarginRatio;
    return observation.boundingBox.left >= marginX &&
        observation.boundingBox.top >= marginY &&
        observation.boundingBox.right <= observation.imageWidth - marginX &&
        observation.boundingBox.bottom <= observation.imageHeight - marginY;
  }

  bool _isPoseCenter(FaceLivenessObservation observation) {
    return _normalizedYaw(observation).abs() <= _config.centerYawThreshold;
  }

  void _handleFaceLost(DateTime now, String message) {
    final lastSeen = _lastValidFaceAt;
    final active =
        _state.status == LivenessResultStatus.stabilizing ||
        _state.status == LivenessResultStatus.executingAction ||
        _state.status == LivenessResultStatus.returningCenter;
    if (active &&
        lastSeen != null &&
        now.difference(lastSeen) > _config.faceLostTimeout) {
      _fail('Verifikasi keaktifan gagal. Silakan coba lagi.');
      return;
    }
    if (!active) {
      _stabilizingSince = null;
      _setState(
        _state.copyWith(
          status: LivenessResultStatus.waitingForFace,
          message: message,
          stepState: LivenessStepState.waitingForFace,
        ),
      );
    }
  }

  void _handleMultipleFaces(DateTime now) {
    final active =
        _state.status == LivenessResultStatus.stabilizing ||
        _state.status == LivenessResultStatus.executingAction ||
        _state.status == LivenessResultStatus.returningCenter;
    if (!active) {
      _setState(
        _state.copyWith(
          status: LivenessResultStatus.waitingForFace,
          message: 'Pastikan hanya satu wajah di kamera.',
          stepState: LivenessStepState.waitingForFace,
        ),
      );
      return;
    }

    _multipleFaceSince ??= now;
    if (now.difference(_multipleFaceSince!) > _config.multipleFaceGrace) {
      _fail('Verifikasi keaktifan gagal. Silakan coba lagi.');
    }
  }

  void _beginChallenge(DateTime now, FaceLivenessObservation observation) {
    final challenge = LivenessChallenge.random(random: _random);
    _challengeStartedAt = now;
    _stepStartedAt = now;
    _lastValidFaceAt = now;
    _trackingId = observation.trackingId;
    _trackingMismatchSince = null;
    _clearActionStability();
    _blinkPhase = _BlinkPhase.waitingOpen;
    _setState(
      LivenessResult(
        status: LivenessResultStatus.executingAction,
        message: challenge.actions.first.instruction,
        challenge: challenge,
        stepState: LivenessStepState.executingAction,
      ),
    );
    _debugObservation(now, observation);
  }

  bool _hasTimedOut(DateTime now) {
    final started = _challengeStartedAt;
    final stepStarted = _stepStartedAt;
    return (started != null &&
            now.difference(started) > _config.challengeTimeout) ||
        (stepStarted != null &&
            now.difference(stepStarted) > _config.actionTimeout);
  }

  bool _hasSameTracking(DateTime now, FaceLivenessObservation observation) {
    final expected = _trackingId;
    final current = observation.trackingId;
    if (expected == null || current == null) {
      _trackingId ??= current;
      _trackingMismatchSince = null;
      return true;
    }
    if (expected == current) {
      _trackingMismatchSince = null;
      return true;
    }

    _trackingMismatchSince ??= now;
    if (now.difference(_trackingMismatchSince!) >
        _config.trackingMismatchGrace) {
      _fail('Verifikasi keaktifan gagal. Silakan coba lagi.');
    }
    return false;
  }

  bool _matchesAction(
    LivenessAction action,
    FaceLivenessObservation observation,
    DateTime now,
  ) {
    switch (action) {
      case LivenessAction.blink:
        return _matchesBlink(observation);
      case LivenessAction.turnLeft:
        return _matchesStablePose(
          now,
          _normalizedYaw(observation) <= -_config.turnYawThreshold,
        );
      case LivenessAction.turnRight:
        return _matchesStablePose(
          now,
          _normalizedYaw(observation) >= _config.turnYawThreshold,
        );
      case LivenessAction.returnCenter:
        return _matchesStablePose(now, _isPoseCenter(observation));
    }
  }

  bool _matchesStablePose(DateTime now, bool isMatched) {
    if (!isMatched) {
      final lastMatch = _lastActionMatchAt;
      if (lastMatch != null &&
          now.difference(lastMatch) <= _config.noisyFrameGrace) {
        return false;
      }
      _clearActionStability();
      return false;
    }

    _actionStableSince ??= now;
    _lastActionMatchAt = now;
    return now.difference(_actionStableSince!) >= _config.actionStableDuration;
  }

  bool _matchesBlink(FaceLivenessObservation observation) {
    final left = observation.leftEyeOpenProbability;
    final right = observation.rightEyeOpenProbability;
    if (left == null || right == null) {
      return false;
    }
    final open =
        left >= _config.openEyeThreshold && right >= _config.openEyeThreshold;
    final closed =
        left <= _config.closedEyeThreshold &&
        right <= _config.closedEyeThreshold;
    switch (_blinkPhase) {
      case _BlinkPhase.waitingOpen:
        if (open) {
          _blinkPhase = _BlinkPhase.waitingClosed;
        }
        return false;
      case _BlinkPhase.waitingClosed:
        if (closed) {
          _blinkPhase = _BlinkPhase.waitingReopen;
        }
        return false;
      case _BlinkPhase.waitingReopen:
        return open;
    }
  }

  double _normalizedYaw(FaceLivenessObservation observation) {
    return _cameraOrientation.normalizeYaw(observation.headEulerAngleY ?? 0);
  }

  void _completeStep(DateTime now, LivenessAction action) {
    final nextIndex = _state.currentStepIndex + 1;
    final challenge = _state.challenge;
    if (challenge == null) {
      return;
    }
    if (nextIndex >= challenge.actions.length) {
      if (action == LivenessAction.returnCenter) {
        _completeChallenge();
      } else {
        _beginReturnCenter(now, challenge);
      }
      return;
    }
    _stepStartedAt = now;
    _clearActionStability();
    _blinkPhase = _BlinkPhase.waitingOpen;
    final nextAction = challenge.actions[nextIndex];
    _setState(
      _state.copyWith(
        status: nextAction == LivenessAction.returnCenter
            ? LivenessResultStatus.returningCenter
            : LivenessResultStatus.executingAction,
        currentStepIndex: nextIndex,
        message: nextAction.instruction,
        stepState: nextAction == LivenessAction.returnCenter
            ? LivenessStepState.returningCenter
            : LivenessStepState.executingAction,
      ),
    );
  }

  void _beginReturnCenter(DateTime now, LivenessChallenge challenge) {
    _stepStartedAt = now;
    _clearActionStability();
    _blinkPhase = _BlinkPhase.waitingOpen;
    _setState(
      _state.copyWith(
        status: LivenessResultStatus.returningCenter,
        currentStepIndex: challenge.actions.length,
        message: LivenessAction.returnCenter.instruction,
        stepState: LivenessStepState.returningCenter,
      ),
    );
  }

  void _completeChallenge() {
    _setState(
      _state.copyWith(
        status: LivenessResultStatus.completed,
        message: 'Memverifikasi wajah...',
        stepState: LivenessStepState.passed,
      ),
    );
  }

  void _resetSession({
    required LivenessResultStatus status,
    required String message,
    required LivenessStepState stepState,
  }) {
    _challengeStartedAt = null;
    _stepStartedAt = null;
    _lastValidFaceAt = null;
    _stabilizingSince = null;
    _multipleFaceSince = null;
    _trackingMismatchSince = null;
    _trackingId = null;
    _clearActionStability();
    _blinkPhase = _BlinkPhase.waitingOpen;
    _isSubmitting = false;
    _setState(
      LivenessResult(status: status, message: message, stepState: stepState),
    );
  }

  void _clearActionStability() {
    _actionStableSince = null;
    _lastActionMatchAt = null;
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
    _setState(
      _state.copyWith(
        status: LivenessResultStatus.failure,
        message: error.message,
        sessionExpired: error.kind == FaceFailureKind.sessionExpired,
      ),
    );
    _isSubmitting = false;
  }

  void _fail(String message) {
    _setState(
      _state.copyWith(
        status: LivenessResultStatus.failure,
        message: message,
        stepState: LivenessStepState.failed,
      ),
    );
    _isSubmitting = false;
  }

  void _debugObservation(DateTime now, FaceLivenessObservation observation) {
    if (!kDebugMode) {
      return;
    }
    final action = _state.status == LivenessResultStatus.returningCenter
        ? LivenessAction.returnCenter
        : _state.currentAction;
    final elapsed = _stepStartedAt == null
        ? 0
        : now.difference(_stepStartedAt!).inMilliseconds;
    final stable = _actionStableSince == null
        ? 0
        : now.difference(_actionStableSince!).inMilliseconds;
    debugPrint(
      'liveness action=$action yaw=${_normalizedYaw(observation).toStringAsFixed(1)} '
      'leftEye=${observation.leftEyeOpenProbability?.toStringAsFixed(2)} '
      'rightEye=${observation.rightEyeOpenProbability?.toStringAsFixed(2)} '
      'stableMs=$stable elapsedMs=$elapsed trackingId=${observation.trackingId}',
    );
  }

  void _setState(LivenessResult state) {
    _state = state;
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
