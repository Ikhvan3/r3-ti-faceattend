import 'dart:async';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../data/face_detector_service.dart';
import '../domain/face_liveness_models.dart';
import 'face_camera_capture_page.dart';
import 'face_camera_preview.dart';
import 'face_liveness_controller.dart';

class FaceLivenessPage extends StatelessWidget {
  const FaceLivenessPage({super.key});

  @override
  Widget build(BuildContext context) {
    return FaceCameraCapturePage(
      title: 'Uji Keaktifan Wajah',
      permissionMessage: 'Izin kamera diperlukan untuk uji keaktifan wajah.',
      imageFormatGroup: ImageFormatGroup.yuv420,
      builder: (context, cameraController, capture) => _LivenessContent(
        cameraController: cameraController,
        onCapture: capture,
      ),
    );
  }
}

class _LivenessContent extends StatefulWidget {
  const _LivenessContent({
    required this.cameraController,
    required this.onCapture,
  });

  final CameraController cameraController;
  final CameraImageCapture onCapture;

  @override
  State<_LivenessContent> createState() => _LivenessContentState();
}

class _LivenessContentState extends State<_LivenessContent> {
  bool _isProcessingFrame = false;
  bool _verificationStarted = false;
  bool _disposed = false;
  DateTime? _lastFrameStartedAt;
  late final FaceDetectorService _detector;

  FaceLivenessController get _controller =>
      context.read<FaceLivenessController>();

  @override
  void initState() {
    super.initState();
    _detector = context.read<FaceDetectorService>();
    _controller.addListener(_handleControllerState);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        _startChallenge();
      }
    });
  }

  @override
  void dispose() {
    _disposed = true;
    _controller.removeListener(_handleControllerState);
    unawaited(_stopImageStream());
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<FaceLivenessController>().state;
    final isFailure = state.status == LivenessResultStatus.failure;
    final isSuccess = state.status == LivenessResultStatus.success;
    final progress = state.progress;

    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        FaceCameraPreview(controller: widget.cameraController),
        const SizedBox(height: 16),
        LinearProgressIndicator(value: progress == 0 ? null : progress),
        const SizedBox(height: 12),
        Text(
          state.message,
          textAlign: TextAlign.center,
          style: isSuccess
              ? TextStyle(
                  color: state.verified
                      ? Colors.green.shade700
                      : Colors.red.shade700,
                  fontWeight: FontWeight.w700,
                )
              : isFailure
              ? TextStyle(color: Colors.red.shade700)
              : null,
        ),
        const SizedBox(height: 8),
        Text(
          state.totalSteps == 0
              ? 'Menunggu wajah'
              : '${state.completedSteps}/${state.totalSteps}',
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: _canRetry(state) ? _startChallenge : null,
          icon: const Icon(Icons.refresh),
          label: const Text('Coba Lagi'),
        ),
      ],
    );
  }

  bool _canRetry(LivenessResult state) {
    return state.status == LivenessResultStatus.failure ||
        state.status == LivenessResultStatus.success;
  }

  Future<void> _startChallenge() async {
    _verificationStarted = false;
    _controller.start();
    await _startImageStream();
  }

  Future<void> _startImageStream() async {
    if (widget.cameraController.value.isStreamingImages) {
      return;
    }
    await widget.cameraController.startImageStream(_processFrame);
  }

  Future<void> _stopImageStream() async {
    if (!widget.cameraController.value.isStreamingImages) {
      return;
    }
    await widget.cameraController.stopImageStream();
  }

  Future<void> _processFrame(CameraImage image) async {
    if (_isProcessingFrame || _disposed) {
      return;
    }
    final now = DateTime.now();
    final lastStarted = _lastFrameStartedAt;
    if (lastStarted != null &&
        now.difference(lastStarted) < const LivenessConfig().frameThrottle) {
      return;
    }
    _lastFrameStartedAt = now;
    _isProcessingFrame = true;
    try {
      final faces = await _detector.detectCameraImage(
        image: image,
        camera: widget.cameraController.description,
        deviceOrientation: widget.cameraController.value.deviceOrientation,
      );
      if (_disposed) {
        return;
      }
      _controller.processFaces(faces);
    } catch (_) {
      _controller.failFromDetector(
        'Kamera belum dapat membaca wajah. Silakan coba lagi.',
      );
    } finally {
      _isProcessingFrame = false;
    }
  }

  void _handleControllerState() {
    final state = _controller.state;
    if (state.status != LivenessResultStatus.completed ||
        _verificationStarted) {
      return;
    }
    _verificationStarted = true;
    unawaited(_verifyAfterLiveness());
  }

  Future<void> _verifyAfterLiveness() async {
    await _stopImageStream();
    await _controller.verifyAfterLiveness(widget.onCapture);
  }
}
