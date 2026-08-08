import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../face/data/face_detector_service.dart';
import '../../face/data/face_embedding_service.dart';
import '../../face/data/face_repository.dart';
import '../../face/presentation/face_liveness_controller.dart';
import '../../face/presentation/face_liveness_page.dart';
import '../domain/attendance_failure.dart';
import '../domain/attendance_models.dart';
import 'attendance_controller.dart';
import 'attendance_formatters.dart';

class AttendanceCard extends StatelessWidget {
  const AttendanceCard({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<AttendanceController>();

    if (controller.status == AttendanceControllerStatus.loading ||
        controller.status == AttendanceControllerStatus.initial) {
      return const _AttendanceLoadingCard();
    }

    if (controller.today == null) {
      return _AttendanceErrorCard(
        message:
            controller.errorMessage ?? 'Data absensi belum dapat ditampilkan.',
        onRetry: controller.refreshToday,
      );
    }

    final today = controller.today!;
    return RefreshIndicator(
      onRefresh: controller.refreshToday,
      child: ListView(
        padding: EdgeInsets.zero,
        shrinkWrap: true,
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      DecoratedBox(
                        decoration: BoxDecoration(
                          color: Theme.of(
                            context,
                          ).colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: const Padding(
                          padding: EdgeInsets.all(10),
                          child: Icon(Icons.fact_check_outlined),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Absensi Hari Ini',
                          style: Theme.of(context).textTheme.titleMedium
                              ?.copyWith(fontWeight: FontWeight.w700),
                        ),
                      ),
                      IconButton(
                        tooltip: 'Refresh',
                        onPressed: controller.isBusy
                            ? null
                            : controller.refreshToday,
                        icon: controller.isRefreshing
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.refresh),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  _StatusPill(state: today.state),
                  const SizedBox(height: 16),
                  _InfoLine(
                    label: 'Tanggal',
                    value: formatAttendanceDate(today.attendanceDate),
                  ),
                  _InfoLine(label: 'Jadwal', value: today.schedule.name),
                  _InfoLine(
                    label: 'Jam kerja',
                    value:
                        '${today.schedule.startTime} - ${today.schedule.endTime}',
                  ),
                  if (today.schedule.graceMinutes > 0)
                    _InfoLine(
                      label: 'Toleransi',
                      value: '${today.schedule.graceMinutes} menit',
                    ),
                  _InfoLine(
                    label: 'Check-in',
                    value: formatAttendanceTime(today.checkInAt),
                  ),
                  _InfoLine(
                    label: 'Check-out',
                    value: formatAttendanceTime(today.checkOutAt),
                  ),
                  if (controller.errorMessage != null) ...[
                    const SizedBox(height: 12),
                    _InlineMessage(
                      message: controller.errorMessage!,
                      icon: Icons.info_outline_rounded,
                    ),
                  ],
                  if (controller.status ==
                      AttendanceControllerStatus.actionLoading) ...[
                    const SizedBox(height: 12),
                    _ActionProgress(step: controller.currentStep),
                  ],
                  const SizedBox(height: 16),
                  _ActionButtons(today: today),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ActionProgress extends StatelessWidget {
  const _ActionProgress({required this.step});

  final AttendanceActionStep? step;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainer,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          children: [
            _ProgressLine(
              icon: Icons.my_location_outlined,
              label: 'Mengambil lokasi perangkat',
              active: step == AttendanceActionStep.location,
              done: _isAfter(AttendanceActionStep.location),
            ),
            const SizedBox(height: 8),
            _ProgressLine(
              icon: Icons.face_retouching_natural_outlined,
              label: 'Memverifikasi keaktifan dan wajah',
              active: step == AttendanceActionStep.face,
              done: _isAfter(AttendanceActionStep.face),
            ),
            const SizedBox(height: 8),
            _ProgressLine(
              icon: Icons.cloud_upload_outlined,
              label: 'Mengirim absensi ke server',
              active: step == AttendanceActionStep.submit,
              done: false,
            ),
          ],
        ),
      ),
    );
  }

  bool _isAfter(AttendanceActionStep current) {
    final active = step;
    if (active == null) {
      return false;
    }
    return active.index > current.index;
  }
}

class _ProgressLine extends StatelessWidget {
  const _ProgressLine({
    required this.icon,
    required this.label,
    required this.active,
    required this.done,
  });

  final IconData icon;
  final String label;
  final bool active;
  final bool done;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = active || done
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;

    return Row(
      children: [
        if (active)
          SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2, color: color),
          )
        else
          Icon(
            done ? Icons.check_circle_rounded : icon,
            size: 20,
            color: color,
          ),
        const SizedBox(width: 10),
        Expanded(
          child: Text(
            label,
            style: theme.textTheme.bodyMedium?.copyWith(color: color),
          ),
        ),
      ],
    );
  }
}

class _InlineMessage extends StatelessWidget {
  const _InlineMessage({required this.message, required this.icon});

  final String message;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.colorScheme.error.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, size: 18, color: theme.colorScheme.error),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                message,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.error,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ActionButtons extends StatelessWidget {
  const _ActionButtons({required this.today});

  final AttendanceToday today;

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<AttendanceController>();
    final isCheckingIn = controller.currentAction == AttendanceAction.checkIn;
    final isCheckingOut = controller.currentAction == AttendanceAction.checkOut;

    switch (today.state) {
      case AttendanceState.notCheckedIn:
        return FilledButton.icon(
          onPressed: today.canCheckIn && !controller.isBusy
              ? () => _confirmCheckIn(context)
              : null,
          icon: isCheckingIn
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.login),
          label: const Text('Check-in'),
        );
      case AttendanceState.checkedIn:
        return FilledButton.icon(
          onPressed: today.canCheckOut && !controller.isBusy
              ? () => _confirmCheckOut(context)
              : null,
          icon: isCheckingOut
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.logout),
          label: const Text('Check-out'),
        );
      case AttendanceState.completed:
        return FilledButton.icon(
          onPressed: null,
          icon: const Icon(Icons.check_circle_outline),
          label: const Text('Absensi selesai'),
        );
    }
  }

  Future<void> _confirmCheckIn(BuildContext context) async {
    final confirmed = await _showAttendanceDialog(
      context: context,
      title: 'Konfirmasi Check-in',
      message:
          'Anda akan melakukan check-in menggunakan lokasi perangkat dan waktu server.',
      confirmLabel: 'Check-in',
    );
    if (!context.mounted || !confirmed) {
      return;
    }

    final controller = context.read<AttendanceController>();
    await controller.checkIn(
      () => _openAttendanceLiveness(context, 'CHECK_IN'),
    );
    if (!context.mounted) {
      return;
    }
    _showFeedback(context, controller, 'Check-in berhasil.');
  }

  Future<void> _confirmCheckOut(BuildContext context) async {
    final confirmed = await _showAttendanceDialog(
      context: context,
      title: 'Konfirmasi Check-out',
      message:
          'Anda akan menyelesaikan absensi hari ini menggunakan lokasi perangkat dan waktu server.',
      confirmLabel: 'Check-out',
    );
    if (!context.mounted || !confirmed) {
      return;
    }

    final controller = context.read<AttendanceController>();
    await controller.checkOut(
      () => _openAttendanceLiveness(context, 'CHECK_OUT'),
    );
    if (!context.mounted) {
      return;
    }
    _showFeedback(context, controller, 'Check-out berhasil.');
  }

  Future<bool> _showAttendanceDialog({
    required BuildContext context,
    required String title,
    required String message,
    required String confirmLabel,
  }) async {
    return await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (dialogContext) {
            return AlertDialog(
              title: Text(title),
              content: Text(message),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(dialogContext).pop(false),
                  child: const Text('Batal'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(dialogContext).pop(true),
                  child: Text(confirmLabel),
                ),
              ],
            );
          },
        ) ??
        false;
  }

  void _showFeedback(
    BuildContext context,
    AttendanceController controller,
    String successMessage,
  ) {
    final message = controller.errorMessage ?? successMessage;
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<String> _openAttendanceLiveness(
    BuildContext context,
    String purpose,
  ) async {
    final grant = await Navigator.of(context).push<String>(
      MaterialPageRoute<String>(
        builder: (_) {
          final detector = MlKitFaceDetectorService.liveness();
          return MultiProvider(
            providers: [
              Provider<FaceDetectorService>.value(value: detector),
              ChangeNotifierProvider<FaceLivenessController>(
                create: (_) => FaceLivenessController(
                  repository: context.read<FaceRepository>(),
                  detector: detector,
                  embeddingService: TfliteFaceEmbeddingService(),
                  attendancePurpose: purpose,
                ),
              ),
            ],
            child: const FaceLivenessPage(),
          );
        },
      ),
    );
    if (grant == null || grant.isEmpty) {
      throw const AttendanceFailure(
        AttendanceFailureKind.faceVerificationRejected,
        'Verifikasi keaktifan gagal.',
      );
    }
    return grant;
  }
}

class _AttendanceLoadingCard extends StatelessWidget {
  const _AttendanceLoadingCard();

  @override
  Widget build(BuildContext context) {
    return const Card(
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(width: 160, child: LinearProgressIndicator()),
            SizedBox(height: 16),
            SizedBox(width: double.infinity, child: LinearProgressIndicator()),
            SizedBox(height: 12),
            SizedBox(width: 220, child: LinearProgressIndicator()),
          ],
        ),
      ),
    );
  }
}

class _AttendanceErrorCard extends StatelessWidget {
  const _AttendanceErrorCard({required this.message, required this.onRetry});

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
              'Absensi belum tersedia',
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

class _InfoLine extends StatelessWidget {
  const _InfoLine({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 92,
            child: Text(label, style: const TextStyle(color: Colors.black54)),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({required this.state});

  final AttendanceState state;

  @override
  Widget build(BuildContext context) {
    return Chip(label: Text(state.label), avatar: Icon(_icon, size: 18));
  }

  IconData get _icon {
    switch (state) {
      case AttendanceState.notCheckedIn:
        return Icons.schedule;
      case AttendanceState.checkedIn:
        return Icons.login;
      case AttendanceState.completed:
        return Icons.check_circle_outline;
    }
  }
}
