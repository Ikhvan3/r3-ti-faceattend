import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../domain/attendance_models.dart';
import 'attendance_controller.dart';
import 'attendance_formatters.dart';
import 'attendance_history_page.dart';

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
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
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
                    Text(
                      controller.errorMessage!,
                      style: TextStyle(color: Colors.red.shade700),
                    ),
                  ],
                  const SizedBox(height: 16),
                  _ActionButtons(today: today),
                  const SizedBox(height: 8),
                  OutlinedButton.icon(
                    onPressed: controller.isBusy
                        ? null
                        : () {
                            Navigator.of(context).push(
                              MaterialPageRoute<void>(
                                builder: (_) => const AttendanceHistoryPage(),
                              ),
                            );
                          },
                    icon: const Icon(Icons.history),
                    label: const Text('Riwayat'),
                  ),
                ],
              ),
            ),
          ),
        ],
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
      message: 'Anda akan melakukan check-in menggunakan waktu server.',
      confirmLabel: 'Check-in',
    );
    if (!context.mounted || !confirmed) {
      return;
    }

    final controller = context.read<AttendanceController>();
    await controller.checkIn();
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
          'Anda akan menyelesaikan absensi hari ini menggunakan waktu server.',
      confirmLabel: 'Check-out',
    );
    if (!context.mounted || !confirmed) {
      return;
    }

    final controller = context.read<AttendanceController>();
    await controller.checkOut();
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
