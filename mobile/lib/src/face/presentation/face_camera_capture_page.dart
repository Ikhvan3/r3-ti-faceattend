import 'package:camera/camera.dart';
import 'package:flutter/material.dart';

import '../domain/face_failure.dart';

typedef CameraImageCapture = Future<String> Function();

class FaceCameraCapturePage extends StatefulWidget {
  const FaceCameraCapturePage({
    required this.title,
    required this.permissionMessage,
    required this.builder,
    super.key,
  });

  final String title;
  final String permissionMessage;
  final Widget Function(
    BuildContext context,
    CameraController cameraController,
    CameraImageCapture capture,
  )
  builder;

  @override
  State<FaceCameraCapturePage> createState() => _FaceCameraCapturePageState();
}

class _FaceCameraCapturePageState extends State<FaceCameraCapturePage>
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
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: FutureBuilder<void>(
        future: _cameraInitialization,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (_cameraError != null || _cameraController == null) {
            return _CameraError(message: _cameraError ?? 'Kamera belum siap.');
          }
          return widget.builder(context, _cameraController!, _captureImage);
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
      return widget.permissionMessage;
    }
    return 'Kamera belum dapat digunakan.';
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
