import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'face_camera_capture_page.dart';
import 'face_camera_preview.dart';
import 'face_enrollment_controller.dart';

class FaceEnrollmentPage extends StatelessWidget {
  const FaceEnrollmentPage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<FaceEnrollmentController>();
    if (controller.status == FaceControllerStatus.success) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (context.mounted) {
          Navigator.of(context).pop();
        }
      });
    }

    return FaceCameraCapturePage(
      title: 'Daftarkan Wajah',
      permissionMessage: 'Izin kamera diperlukan untuk enrollment wajah.',
      builder: (context, cameraController, capture) => _EnrollmentContent(
        cameraController: cameraController,
        onCapture: capture,
      ),
    );
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
        FaceCameraPreview(controller: cameraController),
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
