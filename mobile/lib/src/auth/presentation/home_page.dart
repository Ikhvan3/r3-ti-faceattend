import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../attendance/data/attendance_repository.dart';
import '../../attendance/data/location_service.dart';
import '../../attendance/presentation/attendance_card.dart';
import '../../attendance/presentation/attendance_controller.dart';
import '../../attendance/presentation/attendance_history_page.dart';
import '../../face/data/face_detector_service.dart';
import '../../face/data/face_embedding_service.dart';
import '../../face/data/face_repository.dart';
import '../../face/presentation/face_enrollment_controller.dart';
import '../../face/presentation/face_status_card.dart';
import '../domain/user_profile.dart';
import 'profile_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({required this.user, super.key});

  final UserProfile user;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _selectedIndex = 0;

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider<AttendanceController>(
          create: (context) => AttendanceController(
            context.read<AttendanceRepository>(),
            context.read<LocationService>(),
          )..initialize(),
        ),
        ChangeNotifierProvider<FaceEnrollmentController>(
          create: (context) => FaceEnrollmentController(
            repository: context.read<FaceRepository>(),
            detector: MlKitFaceDetectorService(),
            embeddingService: TfliteFaceEmbeddingService(),
          )..loadStatus(),
        ),
      ],
      child: Scaffold(
        body: IndexedStack(
          index: _selectedIndex,
          children: [
            _AttendanceHomeBody(user: widget.user),
            const AttendanceHistoryPage(embedded: true),
            ProfilePage(user: widget.user, embedded: true),
          ],
        ),
        bottomNavigationBar: NavigationBar(
          selectedIndex: _selectedIndex,
          onDestinationSelected: (index) =>
              setState(() => _selectedIndex = index),
          destinations: const [
            NavigationDestination(
              icon: Icon(Icons.home_outlined),
              selectedIcon: Icon(Icons.home_rounded),
              label: 'Beranda',
            ),
            NavigationDestination(
              icon: Icon(Icons.history_outlined),
              selectedIcon: Icon(Icons.history_rounded),
              label: 'Riwayat',
            ),
            NavigationDestination(
              icon: Icon(Icons.person_outline_rounded),
              selectedIcon: Icon(Icons.person_rounded),
              label: 'Profil',
            ),
          ],
        ),
      ),
    );
  }
}

class _AttendanceHomeBody extends StatefulWidget {
  const _AttendanceHomeBody({required this.user});

  final UserProfile user;

  @override
  State<_AttendanceHomeBody> createState() => _AttendanceHomeBodyState();
}

class _AttendanceHomeBodyState extends State<_AttendanceHomeBody>
    with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<AttendanceController>().refreshToday();
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final now = DateTime.now();

    return CustomScrollView(
      slivers: [
        SliverAppBar(
          pinned: true,
          title: const Text('Beranda'),
          actions: [
            IconButton(
              tooltip: 'Refresh',
              onPressed: () =>
                  context.read<AttendanceController>().refreshToday(),
              icon: const Icon(Icons.refresh_rounded),
            ),
          ],
        ),
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          sliver: SliverList.list(
            children: [
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Halo, ${widget.user.name}',
                          style: theme.textTheme.headlineSmall?.copyWith(
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                        const SizedBox(height: 6),
                        Text(
                          _formatGreetingDate(now),
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                        if (widget.user.position != null) ...[
                          const SizedBox(height: 4),
                          Text(
                            widget.user.position!,
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  CircleAvatar(
                    radius: 24,
                    backgroundColor: theme.colorScheme.primary,
                    foregroundColor: theme.colorScheme.onPrimary,
                    child: Text(_initials(widget.user.name)),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              Row(
                children: [
                  Expanded(
                    child: _InfoTile(
                      icon: Icons.badge_outlined,
                      label: 'NIK',
                      value: widget.user.employeeNumber,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _InfoTile(
                      icon: Icons.verified_user_outlined,
                      label: 'Status',
                      value: _statusLabel(widget.user.accountStatus),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              const AttendanceCard(),
              const SizedBox(height: 18),
              const FaceStatusCard(),
            ],
          ),
        ),
      ],
    );
  }
}

class _InfoTile extends StatelessWidget {
  const _InfoTile({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: theme.colorScheme.primary),
            const SizedBox(height: 10),
            Text(
              label,
              style: theme.textTheme.labelMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              value,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

String _formatGreetingDate(DateTime date) {
  const days = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu'];
  const months = [
    'Januari',
    'Februari',
    'Maret',
    'April',
    'Mei',
    'Juni',
    'Juli',
    'Agustus',
    'September',
    'Oktober',
    'November',
    'Desember',
  ];

  return '${days[date.weekday - 1]}, ${date.day} ${months[date.month - 1]} ${date.year}';
}

String _initials(String name) {
  final parts = name
      .trim()
      .split(RegExp(r'\s+'))
      .where((part) => part.isNotEmpty)
      .toList();
  if (parts.isEmpty) {
    return 'TI';
  }
  if (parts.length == 1) {
    return parts.first.characters.first.toUpperCase();
  }
  return '${parts.first.characters.first}${parts.last.characters.first}'
      .toUpperCase();
}

String _statusLabel(String status) {
  switch (status) {
    case 'ACTIVE':
      return 'aktif';
    case 'INACTIVE':
      return 'nonaktif';
    case 'SUSPENDED':
      return 'ditangguhkan';
    default:
      return 'belum tersedia';
  }
}
