import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../data/attendance_repository.dart';
import '../domain/attendance_failure.dart';
import '../domain/attendance_models.dart';
import 'attendance_formatters.dart';

class AttendanceHistoryPage extends StatefulWidget {
  const AttendanceHistoryPage({super.key});

  @override
  State<AttendanceHistoryPage> createState() => _AttendanceHistoryPageState();
}

class _AttendanceHistoryPageState extends State<AttendanceHistoryPage> {
  static const _pageSize = 10;

  AttendanceHistoryResponse? _history;
  String? _errorMessage;
  bool _isLoading = true;
  int _page = 1;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final repository = context.read<AttendanceRepository>();
      final history = await repository.loadHistory(
        page: _page,
        pageSize: _pageSize,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _history = history;
        _isLoading = false;
      });
    } on AttendanceFailure catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = error.message;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Riwayat Absensi')),
      body: _body(),
    );
  }

  Widget _body() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_errorMessage!, textAlign: TextAlign.center),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: _load,
                icon: const Icon(Icons.refresh),
                label: const Text('Coba Lagi'),
              ),
            ],
          ),
        ),
      );
    }

    final history = _history;
    if (history == null || history.items.isEmpty) {
      return const Center(child: Text('Belum ada riwayat absensi.'));
    }

    return Column(
      children: [
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemBuilder: (context, index) {
              return _HistoryTile(record: history.items[index]);
            },
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemCount: history.items.length,
          ),
        ),
        _PaginationBar(
          pagination: history.pagination,
          onPrevious: history.pagination.page > 1
              ? () {
                  setState(() => _page--);
                  _load();
                }
              : null,
          onNext: history.pagination.page < history.pagination.totalPages
              ? () {
                  setState(() => _page++);
                  _load();
                }
              : null,
        ),
      ],
    );
  }
}

class _HistoryTile extends StatelessWidget {
  const _HistoryTile({required this.record});

  final AttendanceRecord record;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        title: Text(formatAttendanceDate(record.attendanceDate)),
        subtitle: Text(
          '${record.schedule.name}\n'
          'Check-in ${formatAttendanceTime(record.checkInAt)} - '
          'Check-out ${formatAttendanceTime(record.checkOutAt)}',
        ),
        isThreeLine: true,
        trailing: Text(record.state.label),
      ),
    );
  }
}

class _PaginationBar extends StatelessWidget {
  const _PaginationBar({
    required this.pagination,
    required this.onPrevious,
    required this.onNext,
  });

  final AttendancePagination pagination;
  final VoidCallback? onPrevious;
  final VoidCallback? onNext;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            OutlinedButton(
              onPressed: onPrevious,
              child: const Text('Sebelumnya'),
            ),
            Expanded(
              child: Text(
                'Halaman ${pagination.page} dari ${pagination.totalPages}',
                textAlign: TextAlign.center,
              ),
            ),
            OutlinedButton(onPressed: onNext, child: const Text('Berikutnya')),
          ],
        ),
      ),
    );
  }
}
