import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../domain/face_status.dart';
import '../data/face_detector_service.dart';
import '../data/face_embedding_service.dart';
import '../data/face_repository.dart';
import 'face_enrollment_controller.dart';
import 'face_enrollment_page.dart';
import 'face_verification_controller.dart';
import 'face_verification_page.dart';

class FaceStatusCard extends StatelessWidget {
  const FaceStatusCard({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<FaceEnrollmentController>();

    if (controller.status == FaceControllerStatus.loadingStatus ||
        controller.status == FaceControllerStatus.initial) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: LinearProgressIndicator(),
        ),
      );
    }

    final status = controller.faceStatus;
    if (status == null) {
      return _FaceErrorCard(
        message:
            controller.errorMessage ?? 'Status wajah belum dapat ditampilkan.',
        onRetry: controller.loadStatus,
      );
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    'Enrollment Wajah',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                IconButton(
                  tooltip: 'Refresh',
                  onPressed: controller.isBusy ? null : controller.loadStatus,
                  icon: const Icon(Icons.refresh),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Chip(
              avatar: Icon(
                status.enrolled
                    ? Icons.verified_user_outlined
                    : Icons.face_retouching_off_outlined,
                size: 18,
              ),
              label: Text(status.status.label),
            ),
            if (status.enrolledAt != null) ...[
              const SizedBox(height: 8),
              Text('Terdaftar: ${_formatDateTime(status.enrolledAt!)}'),
            ],
            if (controller.errorMessage != null) ...[
              const SizedBox(height: 8),
              Text(
                controller.errorMessage!,
                style: TextStyle(color: Colors.red.shade700),
              ),
            ],
            const SizedBox(height: 12),
            if (status.status == FaceEnrollmentStatus.notEnrolled)
              FilledButton.icon(
                onPressed: controller.isBusy
                    ? null
                    : () => _openEnrollment(context),
                icon: const Icon(Icons.face_retouching_natural),
                label: const Text('Daftarkan Wajah'),
              )
            else ...[
              FilledButton.icon(
                onPressed: controller.isBusy
                    ? null
                    : () => _openVerification(context),
                icon: const Icon(Icons.verified_user_outlined),
                label: const Text('Uji Verifikasi Wajah'),
              ),
              const SizedBox(height: 8),
              OutlinedButton.icon(
                onPressed: controller.isBusy
                    ? null
                    : () => _confirmReset(context),
                icon: const Icon(Icons.restart_alt),
                label: const Text('Atur Ulang Wajah'),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Future<void> _openEnrollment(BuildContext context) async {
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => ChangeNotifierProvider<FaceEnrollmentController>.value(
          value: context.read<FaceEnrollmentController>(),
          child: const FaceEnrollmentPage(),
        ),
      ),
    );
    if (context.mounted) {
      await context.read<FaceEnrollmentController>().loadStatus();
    }
  }

  Future<void> _openVerification(BuildContext context) async {
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => ChangeNotifierProvider<FaceVerificationController>(
          create: (_) => FaceVerificationController(
            repository: context.read<FaceRepository>(),
            detector: MlKitFaceDetectorService(),
            embeddingService: TfliteFaceEmbeddingService(),
          ),
          child: const FaceVerificationPage(),
        ),
      ),
    );
  }

  Future<void> _confirmReset(BuildContext context) async {
    final confirmed =
        await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (dialogContext) {
            return AlertDialog(
              title: const Text('Atur Ulang Wajah'),
              content: const Text('Enrollment wajah akan dihapus dari server.'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(dialogContext).pop(false),
                  child: const Text('Batal'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(dialogContext).pop(true),
                  child: const Text('Atur Ulang'),
                ),
              ],
            );
          },
        ) ??
        false;
    if (!context.mounted || !confirmed) {
      return;
    }
    await context.read<FaceEnrollmentController>().resetEnrollment();
  }
}

class _FaceErrorCard extends StatelessWidget {
  const _FaceErrorCard({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Enrollment Wajah',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 8),
            Text(message),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Coba Lagi'),
            ),
          ],
        ),
      ),
    );
  }
}

String _formatDateTime(DateTime value) {
  final local = value.toLocal();
  String two(int number) => number.toString().padLeft(2, '0');
  return '${two(local.day)}/${two(local.month)}/${local.year} ${two(local.hour)}:${two(local.minute)}';
}
