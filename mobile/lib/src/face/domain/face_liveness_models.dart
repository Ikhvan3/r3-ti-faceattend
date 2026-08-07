import 'dart:math';

enum LivenessAction { blink, turnLeft, turnRight, returnCenter }

extension LivenessActionLabel on LivenessAction {
  String get instruction {
    switch (this) {
      case LivenessAction.blink:
        return 'Kedipkan mata.';
      case LivenessAction.turnLeft:
        return 'Hadap sedikit ke kiri.';
      case LivenessAction.turnRight:
        return 'Hadap sedikit ke kanan.';
      case LivenessAction.returnCenter:
        return 'Kembali menghadap depan.';
    }
  }
}

enum LivenessStepState {
  waitingForFace,
  stabilizing,
  executingAction,
  returningCenter,
  passed,
  failed,
}

enum LivenessResultStatus {
  initializing,
  waitingForFace,
  stabilizing,
  executingAction,
  returningCenter,
  completed,
  verifying,
  success,
  failure,
}

class LivenessChallenge {
  LivenessChallenge({required List<LivenessAction> actions})
    : actions = List.unmodifiable(actions);

  factory LivenessChallenge.random({
    Random? random,
    int minActions = 2,
    int maxActions = 3,
  }) {
    final rng = random ?? Random.secure();
    final target = minActions + rng.nextInt(maxActions - minActions + 1);
    final actions = <LivenessAction>[];
    var lastTurnNeedsCenter = false;

    while (actions.length < target) {
      final candidates = lastTurnNeedsCenter
          ? const [LivenessAction.returnCenter]
          : const [
              LivenessAction.blink,
              LivenessAction.turnLeft,
              LivenessAction.turnRight,
            ];
      final next = candidates[rng.nextInt(candidates.length)];
      actions.add(next);
      lastTurnNeedsCenter =
          next == LivenessAction.turnLeft || next == LivenessAction.turnRight;
    }

    if (actions.any(
      (action) =>
          action == LivenessAction.turnLeft ||
          action == LivenessAction.turnRight,
    )) {
      final last = actions.last;
      if (last == LivenessAction.turnLeft || last == LivenessAction.turnRight) {
        actions[actions.length - 1] = LivenessAction.returnCenter;
      }
    }

    return LivenessChallenge(actions: actions);
  }

  final List<LivenessAction> actions;
}

class LivenessConfig {
  const LivenessConfig({
    this.openEyeThreshold = 0.60,
    this.closedEyeThreshold = 0.40,
    this.turnYawThreshold = 13,
    this.centerYawThreshold = 10,
    this.maxRollDegrees = 25,
    this.minFaceBoxRatio = 0.20,
    this.centerOffsetRatio = 0.24,
    this.actionTimeout = const Duration(seconds: 10),
    this.challengeTimeout = const Duration(seconds: 40),
    this.faceLostTimeout = const Duration(milliseconds: 2200),
    this.multipleFaceGrace = const Duration(milliseconds: 1200),
    this.trackingMismatchGrace = const Duration(milliseconds: 1800),
    this.stabilizationDuration = const Duration(milliseconds: 550),
    this.actionStableDuration = const Duration(milliseconds: 140),
    this.noisyFrameGrace = const Duration(milliseconds: 500),
    this.frameThrottle = const Duration(milliseconds: 80),
    this.sampleInterval = const Duration(milliseconds: 650),
  });

  final double openEyeThreshold;
  final double closedEyeThreshold;
  final double turnYawThreshold;
  final double centerYawThreshold;
  final double maxRollDegrees;
  final double minFaceBoxRatio;
  final double centerOffsetRatio;
  final Duration actionTimeout;
  final Duration challengeTimeout;
  final Duration faceLostTimeout;
  final Duration multipleFaceGrace;
  final Duration trackingMismatchGrace;
  final Duration stabilizationDuration;
  final Duration actionStableDuration;
  final Duration noisyFrameGrace;
  final Duration frameThrottle;
  final Duration sampleInterval;
}

class LivenessResult {
  const LivenessResult({
    required this.status,
    required this.message,
    this.challenge,
    this.currentStepIndex = 0,
    this.stepState = LivenessStepState.waitingForFace,
    this.verified = false,
    this.sessionExpired = false,
  });

  final LivenessResultStatus status;
  final String message;
  final LivenessChallenge? challenge;
  final int currentStepIndex;
  final LivenessStepState stepState;
  final bool verified;
  final bool sessionExpired;

  LivenessAction? get currentAction {
    final activeChallenge = challenge;
    if (activeChallenge == null ||
        currentStepIndex >= activeChallenge.actions.length) {
      return null;
    }
    return activeChallenge.actions[currentStepIndex];
  }

  int get completedSteps => currentStepIndex;
  int get totalSteps => challenge?.actions.length ?? 0;

  double? get progress {
    final total = totalSteps;
    if (total == 0) {
      return null;
    }
    return completedSteps / total;
  }

  LivenessResult copyWith({
    LivenessResultStatus? status,
    String? message,
    LivenessChallenge? challenge,
    int? currentStepIndex,
    LivenessStepState? stepState,
    bool? verified,
    bool? sessionExpired,
  }) {
    return LivenessResult(
      status: status ?? this.status,
      message: message ?? this.message,
      challenge: challenge ?? this.challenge,
      currentStepIndex: currentStepIndex ?? this.currentStepIndex,
      stepState: stepState ?? this.stepState,
      verified: verified ?? this.verified,
      sessionExpired: sessionExpired ?? this.sessionExpired,
    );
  }
}
