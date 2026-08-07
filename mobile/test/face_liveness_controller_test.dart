import 'dart:math';
import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/face/data/face_repository.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_detection_result.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_failure.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_liveness_models.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_model_config.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_verification_result.dart';
import 'package:r3_ti_faceattend/src/face/presentation/face_liveness_controller.dart';

import 'face_test_fakes.dart';

const livenessTestConfig = LivenessConfig(
  stabilizationDuration: Duration(milliseconds: 800),
  actionStableDuration: Duration(milliseconds: 250),
  actionTimeout: Duration(seconds: 6),
  challengeTimeout: Duration(seconds: 25),
  faceLostTimeout: Duration(milliseconds: 900),
  multipleFaceGrace: Duration(milliseconds: 700),
  trackingMismatchGrace: Duration(milliseconds: 500),
  sampleInterval: Duration.zero,
);

void main() {
  test('challenge belum dimulai sebelum stabilization selesai', () {
    final clock = FakeClock();
    final controller = newLivenessController(now: clock.call);
    controller.start();

    controller.processFaces([liveFace()]);
    expect(controller.state.status, LivenessResultStatus.stabilizing);

    clock.advance(const Duration(milliseconds: 799));
    controller.processFaces([liveFace()]);
    expect(controller.state.status, LivenessResultStatus.stabilizing);

    clock.advance(const Duration(milliseconds: 1));
    controller.processFaces([liveFace()]);
    expect(controller.state.status, LivenessResultStatus.executingAction);
  });

  test('timer action dimulai setelah stabilization, bukan initialization', () {
    final clock = FakeClock();
    final controller = newLivenessController(now: clock.call);
    controller.start();

    clock.advance(const Duration(seconds: 10));
    controller.processFaces([liveFace()]);

    expect(controller.state.status, LivenessResultStatus.stabilizing);
    expect(controller.state.status, isNot(LivenessResultStatus.failure));
  });

  test(
    'instruction belum dilakukan dan single noisy frame tidak langsung fail',
    () {
      final clock = FakeClock();
      final controller = newLivenessController(
        random: SequenceRandom([0, 1, 0]),
        now: clock.call,
      );
      startChallenge(controller, clock);

      controller.processFaces([liveFace(headEulerAngleY: 0)]);
      expect(controller.state.currentStepIndex, 0);
      expect(controller.state.status, LivenessResultStatus.executingAction);

      controller.processFaces([]);
      expect(controller.state.status, LivenessResultStatus.executingAction);

      controller.processFaces([liveFace(), liveFace()]);
      expect(controller.state.status, LivenessResultStatus.executingAction);
    },
  );

  test('turn threshold satu frame belum pass, stable beberapa frame pass', () {
    final clock = FakeClock();
    final controller = newLivenessController(
      random: SequenceRandom([0, 1, 0]),
      now: clock.call,
    );
    startChallenge(controller, clock);

    expect(controller.state.currentAction, LivenessAction.turnLeft);
    controller.processFaces([liveFace(headEulerAngleY: 18)]);
    expect(controller.state.currentStepIndex, 0);

    clock.advance(const Duration(milliseconds: 250));
    controller.processFaces([liveFace(headEulerAngleY: 18)]);

    expect(controller.state.currentStepIndex, 1);
    expect(controller.state.status, LivenessResultStatus.returningCenter);
  });

  test('return center harus stable sebelum pass', () {
    final clock = FakeClock();
    final controller = newLivenessController(
      random: SequenceRandom([0, 1, 0]),
      now: clock.call,
    );
    startChallenge(controller, clock);
    passTurnLeft(controller, clock);

    controller.processFaces([liveFace(headEulerAngleY: 0)]);
    expect(controller.state.status, LivenessResultStatus.returningCenter);

    clock.advance(const Duration(milliseconds: 250));
    controller.processFaces([liveFace(headEulerAngleY: 0)]);

    expect(controller.state.status, LivenessResultStatus.completed);
  });

  test(
    'blink open closed open pass, closed saja dan probability null tidak pass',
    () {
      final clock = FakeClock();
      final controller = newLivenessController(
        random: SequenceRandom([0, 0, 0]),
        now: clock.call,
      );
      startChallenge(controller, clock);

      controller.processFaces([liveFace(leftEye: 0.1, rightEye: 0.1)]);
      expect(controller.state.currentStepIndex, 0);

      controller.processFaces([liveFace(leftEye: null, rightEye: null)]);
      expect(controller.state.status, LivenessResultStatus.executingAction);

      controller.processFaces([liveFace()]);
      controller.processFaces([liveFace(leftEye: 0.1, rightEye: 0.1)]);
      controller.processFaces([liveFace()]);

      expect(controller.state.currentStepIndex, 1);
    },
  );

  test('wrong yaw direction tidak pass tetapi tidak langsung fail', () {
    final clock = FakeClock();
    final controller = newLivenessController(
      random: SequenceRandom([0, 1, 0]),
      now: clock.call,
    );
    startChallenge(controller, clock);

    controller.processFaces([liveFace(headEulerAngleY: -20)]);
    clock.advance(const Duration(milliseconds: 250));
    controller.processFaces([liveFace(headEulerAngleY: -20)]);

    expect(controller.state.currentStepIndex, 0);
    expect(controller.state.status, LivenessResultStatus.executingAction);
  });

  test('face lost satu frame tidak fail, melewati grace menjadi failure', () {
    final clock = FakeClock();
    final controller = newLivenessController(now: clock.call);
    startChallenge(controller, clock);

    controller.processFaces([]);
    expect(controller.state.status, isNot(LivenessResultStatus.failure));

    clock.advance(const Duration(milliseconds: 901));
    controller.processFaces([]);
    expect(controller.state.status, LivenessResultStatus.failure);
  });

  test('multiple face persistent dan tracking mismatch persistent fail', () {
    final clock = FakeClock();
    final multiple = newLivenessController(now: clock.call);
    startChallenge(multiple, clock);
    multiple.processFaces([liveFace(), liveFace()]);
    expect(multiple.state.status, isNot(LivenessResultStatus.failure));

    clock.advance(const Duration(milliseconds: 701));
    multiple.processFaces([liveFace(), liveFace()]);
    expect(multiple.state.status, LivenessResultStatus.failure);

    final trackingClock = FakeClock();
    final tracking = newLivenessController(now: trackingClock.call);
    startChallenge(tracking, trackingClock);
    tracking.processFaces([liveFace(trackingId: 2)]);
    expect(tracking.state.status, isNot(LivenessResultStatus.failure));

    trackingClock.advance(const Duration(milliseconds: 501));
    tracking.processFaces([liveFace(trackingId: 2)]);
    expect(tracking.state.status, LivenessResultStatus.failure);
  });

  test('action timeout menghasilkan failure', () {
    final clock = FakeClock();
    final controller = newLivenessController(
      random: SequenceRandom([0, 1, 0]),
      now: clock.call,
    );
    startChallenge(controller, clock);

    clock.advance(const Duration(milliseconds: 6001));
    controller.processFaces([liveFace(headEulerAngleY: 0)]);

    expect(controller.state.status, LivenessResultStatus.failure);
  });

  test('challenge random returns valid action lengths', () {
    final seen = <List<LivenessAction>>[];
    for (var i = 0; i < 20; i += 1) {
      final challenge = LivenessChallenge.random(random: Random(i));
      seen.add(challenge.actions);
      expect(challenge.actions.length, inInclusiveRange(2, 3));
      expect(challenge.actions, isNotEmpty);
    }
    expect(seen.toSet().length, greaterThan(1));
  });

  test('liveness pass then face verify true and false', () async {
    final verifiedApi = FakeFaceApi(status: enrolledStatus);
    final verifiedClock = FakeClock();
    final verified = newLivenessController(
      api: verifiedApi,
      now: verifiedClock.call,
    );
    passCurrentChallenge(verified, verifiedClock);

    await verified.verifyAfterLiveness(() async => 'sample.jpg');

    expect(verifiedApi.verifyCount, 1);
    expect(verified.state.status, LivenessResultStatus.success);
    expect(verified.state.verified, isTrue);
    expect(verified.state.message, 'Verifikasi wajah berhasil.');

    final rejectedApi = FakeFaceApi(
      status: enrolledStatus,
      verifyResult: const FaceVerificationResult(verified: false),
    );
    final rejectedClock = FakeClock();
    final rejected = newLivenessController(
      api: rejectedApi,
      now: rejectedClock.call,
    );
    passCurrentChallenge(rejected, rejectedClock);

    await rejected.verifyAfterLiveness(() async => 'sample.jpg');

    expect(rejected.state.verified, isFalse);
    expect(
      rejected.state.message,
      'Wajah tidak sesuai dengan data yang terdaftar.',
    );
  });

  test(
    'verification API error, retry, and multiple submit are handled',
    () async {
      final api = FakeFaceApi(
        status: enrolledStatus,
        verifyError: const FaceFailure(
          FaceFailureKind.apiUnavailable,
          'offline',
        ),
      );
      final clock = FakeClock();
      final controller = newLivenessController(api: api, now: clock.call);
      passCurrentChallenge(controller, clock);

      await controller.verifyAfterLiveness(() async => 'sample.jpg');
      expect(controller.state.status, LivenessResultStatus.failure);
      expect(controller.state.message, 'offline');

      controller.retry();
      expect(controller.state.status, LivenessResultStatus.waitingForFace);

      final submitApi = FakeFaceApi(status: enrolledStatus);
      final submitClock = FakeClock();
      final submit = newLivenessController(
        api: submitApi,
        now: submitClock.call,
      );
      passCurrentChallenge(submit, submitClock);
      final first = submit.verifyAfterLiveness(() async => 'sample.jpg');
      await submit.verifyAfterLiveness(() async => 'ignored.jpg');
      await first;

      expect(submitApi.verifyCount, 1);
    },
  );
}

FaceLivenessController newLivenessController({
  FakeFaceApi? api,
  Random? random,
  DateTime Function()? now,
  LivenessConfig config = livenessTestConfig,
}) {
  return FaceLivenessController(
    repository: FaceRepository(api: api ?? FakeFaceApi(status: enrolledStatus)),
    detector: FakeFaceDetector(faces: [liveFace()]),
    embeddingService: FakeFaceEmbeddingService(
      embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 1),
    ),
    config: config,
    random: random,
    now: now,
  );
}

FaceDetectionResult liveFace({
  Rect box = const Rect.fromLTWH(80, 80, 220, 220),
  int? trackingId = 1,
  double? leftEye = 0.9,
  double? rightEye = 0.9,
  double? headEulerAngleY,
  double? headEulerAngleZ,
}) {
  return FaceDetectionResult(
    boundingBox: box,
    imageWidth: 400,
    imageHeight: 400,
    trackingId: trackingId,
    leftEyeOpenProbability: leftEye,
    rightEyeOpenProbability: rightEye,
    headEulerAngleY: headEulerAngleY,
    headEulerAngleZ: headEulerAngleZ,
  );
}

void startChallenge(FaceLivenessController controller, FakeClock clock) {
  controller.start();
  controller.processFaces([liveFace()]);
  clock.advance(const Duration(milliseconds: 800));
  controller.processFaces([liveFace()]);
  expect(controller.state.status, LivenessResultStatus.executingAction);
}

void passTurnLeft(FaceLivenessController controller, FakeClock clock) {
  controller.processFaces([liveFace(headEulerAngleY: 18)]);
  clock.advance(const Duration(milliseconds: 250));
  controller.processFaces([liveFace(headEulerAngleY: 18)]);
  expect(controller.state.status, LivenessResultStatus.returningCenter);
}

void passCurrentChallenge(FaceLivenessController controller, FakeClock clock) {
  startChallenge(controller, clock);
  while (controller.state.status == LivenessResultStatus.executingAction ||
      controller.state.status == LivenessResultStatus.returningCenter) {
    final action =
        controller.state.status == LivenessResultStatus.returningCenter
        ? LivenessAction.returnCenter
        : controller.state.currentAction!;
    switch (action) {
      case LivenessAction.blink:
        controller.processFaces([liveFace()]);
        controller.processFaces([liveFace(leftEye: 0.1, rightEye: 0.1)]);
        controller.processFaces([liveFace()]);
      case LivenessAction.turnLeft:
        controller.processFaces([liveFace(headEulerAngleY: 18)]);
        clock.advance(const Duration(milliseconds: 250));
        controller.processFaces([liveFace(headEulerAngleY: 18)]);
      case LivenessAction.turnRight:
        controller.processFaces([liveFace(headEulerAngleY: -18)]);
        clock.advance(const Duration(milliseconds: 250));
        controller.processFaces([liveFace(headEulerAngleY: -18)]);
      case LivenessAction.returnCenter:
        controller.processFaces([liveFace(headEulerAngleY: 0)]);
        clock.advance(const Duration(milliseconds: 250));
        controller.processFaces([liveFace(headEulerAngleY: 0)]);
    }
  }
  expect(controller.state.status, LivenessResultStatus.completed);
}

class FakeClock {
  DateTime _now = DateTime(2026, 8, 7);

  DateTime call() => _now;

  void advance(Duration duration) {
    _now = _now.add(duration);
  }
}

class SequenceRandom implements Random {
  SequenceRandom(this.values);

  final List<int> values;
  int _index = 0;

  @override
  bool nextBool() => nextInt(2) == 0;

  @override
  double nextDouble() => nextInt(1000) / 1000;

  @override
  int nextInt(int max) {
    final value = values[_index % values.length] % max;
    _index += 1;
    return value;
  }
}
