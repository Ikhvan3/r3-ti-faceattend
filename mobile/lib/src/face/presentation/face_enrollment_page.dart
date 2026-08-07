import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../domain/face_failure.dart';
import 'face_enrollment_controller.dart';

class FaceEnrollmentPage extends StatefulWidget {
  const FaceEnrollmentPage({super.key});

  @override
  State<FaceEnrollmentPage> createState() => _FaceEnrollmentPageState();
}

class _FaceEnrollmentPageState extends State<FaceEnrollmentPage>
    with WidgetsBindingObserver {
  CameraController? _cameraController;
  Future<void>? _cameraInitialization;
  String? _cameraError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _cameraInitialization = _initializeCamera();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _disposeCamera();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.inactive ||
        state == AppLifecycleState.paused ||
        state == AppLifecycleState.detached) {
      _disposeCamera();
      return;
    }
    if (state == AppLifecycleState.resumed && _cameraController == null) {
      setState(() {
        _cameraInitialization = _initializeCamera();
      });
    }
  }

  Future<void> _initializeCamera() async {
    setState(() {
      _cameraError = null;
    });
    try {
      final cameras = await availableCameras();
      if (cameras.isEmpty) {
        throw const FaceFailure(
          FaceFailureKind.cameraUnavailable,
          'Kamera belum tersedia.',
        );
      }
      final selected = cameras.firstWhere(
        (camera) => camera.lensDirection == CameraLensDirection.front,
        orElse: () => cameras.first,
      );
      final controller = CameraController(
        selected,
        ResolutionPreset.medium,
        enableAudio: false,
        imageFormatGroup: ImageFormatGroup.jpeg,
      );
      await controller.initialize();
      if (!mounted) {
        await controller.dispose();
        return;
      }
      _cameraController = controller;
    } on CameraException catch (error) {
      _cameraError = _cameraMessage(error);
    } on FaceFailure catch (error) {
      _cameraError = error.message;
    }
  }

  Future<void> _disposeCamera() async {
    final controller = _cameraController;
    _cameraController = null;
    await controller?.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<FaceEnrollmentController>();
    if (controller.status == FaceControllerStatus.success) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          Navigator.of(context).pop();
        }
      });
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Daftarkan Wajah')),
      body: FutureBuilder<void>(
        future: _cameraInitialization,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (_cameraError != null || _cameraController == null) {
            return _CameraError(message: _cameraError ?? 'Kamera belum siap.');
          }
          return _EnrollmentContent(
            cameraController: _cameraController!,
            onCapture: _captureImage,
          );
        },
      ),
    );
  }

  Future<String> _captureImage() async {
    final controller = _cameraController;
    if (controller == null || !controller.value.isInitialized) {
      throw const FaceFailure(
        FaceFailureKind.cameraUnavailable,
        'Kamera belum siap.',
      );
    }
    if (controller.value.isTakingPicture) {
      throw const FaceFailure(
        FaceFailureKind.cameraUnavailable,
        'Kamera sedang memproses sample.',
      );
    }
    try {
      final file = await controller.takePicture();
      return file.path;
    } on CameraException catch (error) {
      throw FaceFailure(
        FaceFailureKind.cameraUnavailable,
        _cameraMessage(error),
      );
    }
  }

  String _cameraMessage(CameraException error) {
    if (error.code == 'CameraAccessDenied' ||
        error.code == 'CameraAccessDeniedWithoutPrompt' ||
        error.code == 'CameraAccessRestricted') {
      return 'Izin kamera diperlukan untuk enrollment wajah.';
    }
    return 'Kamera belum dapat digunakan.';
  }
}

class _EnrollmentContent extends StatelessWidget {
  const _EnrollmentContent({
    required this.cameraController,
    required this.onCapture,
  });

  final CameraController cameraController;
  final FaceImageCapture onCapture;

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<FaceEnrollmentController>();
    final progress = controller.sampleCount / controller.sampleTarget;

    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        AspectRatio(
          aspectRatio: cameraController.value.aspectRatio,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: CameraPreview(cameraController),
          ),
        ),
        const SizedBox(height: 16),
        LinearProgressIndicator(value: progress == 0 ? null : progress),
        const SizedBox(height: 12),
        Text(
          controller.qualityMessage ??
              'Posisikan wajah di tengah frame lalu mulai enrollment.',
          textAlign: TextAlign.center,
        ),
        if (controller.errorMessage != null) ...[
          const SizedBox(height: 8),
          Text(
            controller.errorMessage!,
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.red.shade700),
          ),
        ],
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: controller.isBusy
              ? null
              : () => controller.enrollFromCamera(onCapture),
          icon: controller.isBusy
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.camera_front),
          label: Text(
            controller.isBusy
                ? 'Sample ${controller.sampleCount}/${controller.sampleTarget}'
                : 'Mulai Enrollment',
          ),
        ),
      ],
    );
  }
}

class _CameraError extends StatelessWidget {
  const _CameraError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Text(message, textAlign: TextAlign.center),
      ),
    );
  }
}
