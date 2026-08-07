import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'face_camera_capture_page.dart';
import 'face_verification_controller.dart';

class FaceVerificationPage extends StatelessWidget {
  const FaceVerificationPage({super.key});

  @override
  Widget build(BuildContext context) {
    return FaceCameraCapturePage(
      title: 'Uji Verifikasi Wajah',
      permissionMessage: 'Izin kamera diperlukan untuk verifikasi wajah.',
      builder: (context, cameraController, capture) => _VerificationContent(
        cameraController: cameraController,
        onCapture: capture,
      ),
    );
  }
}

class _VerificationContent extends StatelessWidget {
  const _VerificationContent({
    required this.cameraController,
    required this.onCapture,
  });

  final CameraController cameraController;
  final CameraImageCapture onCapture;

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<FaceVerificationController>();
    final progress = controller.sampleCount / controller.sampleTarget;
    final isSuccess = controller.status == FaceVerificationControllerStatus.success;

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
          controller.message ?? 'Posisikan wajah di tengah frame.',
          textAlign: TextAlign.center,
          style: isSuccess
              ? TextStyle(
                  color: controller.verified
                      ? Colors.green.shade700
                      : Colors.red.shade700,
                  fontWeight: FontWeight.w700,
                )
              : null,
        ),
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: controller.isBusy
              ? null
              : () => controller.verifyFromCamera(onCapture),
          icon: controller.isBusy
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.verified_user_outlined),
          label: Text(
            controller.isBusy
                ? 'Sample ${controller.sampleCount}/${controller.sampleTarget}'
                : 'Mulai Verifikasi',
          ),
        ),
      ],
    );
  }
}
