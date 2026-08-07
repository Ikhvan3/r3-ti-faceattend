import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../attendance/data/attendance_repository.dart';
import '../../attendance/data/location_service.dart';
import '../../attendance/presentation/attendance_card.dart';
import '../../attendance/presentation/attendance_controller.dart';
import '../../face/data/face_detector_service.dart';
import '../../face/data/face_embedding_service.dart';
import '../../face/data/face_repository.dart';
import '../../face/presentation/face_enrollment_controller.dart';
import '../../face/presentation/face_status_card.dart';
import '../domain/user_profile.dart';
import 'auth_controller.dart';
import 'profile_page.dart';

class HomePage extends StatelessWidget {
  const HomePage({required this.user, super.key});

  final UserProfile user;

  @override
  Widget build(BuildContext context) {
    final controller = context.watch<AuthController>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Beranda Pegawai'),
        actions: [
          IconButton(
            tooltip: 'Profil',
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => ProfilePage(user: user),
                ),
              );
            },
            icon: const Icon(Icons.person_outline),
          ),
          IconButton(
            tooltip: 'Logout',
            onPressed: controller.isSubmitting ? null : controller.logout,
            icon: const Icon(Icons.logout),
          ),
        ],
      ),
      body: MultiProvider(
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
        child: _AttendanceHomeBody(user: user),
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
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        Text(
          'Halo, ${widget.user.name}',
          style: Theme.of(
            context,
          ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 8),
        Text('Akun ${_statusLabel(widget.user.accountStatus)}'),
        const SizedBox(height: 20),
        _InfoCard(label: 'Nomor Pegawai', value: widget.user.employeeNumber),
        _InfoCard(label: 'Jabatan', value: widget.user.position ?? '-'),
        _InfoCard(label: 'Email', value: widget.user.email),
        const SizedBox(height: 20),
        const AttendanceCard(),
        const SizedBox(height: 20),
        const FaceStatusCard(),
      ],
    );
  }
}

class _InfoCard extends StatelessWidget {
  const _InfoCard({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(title: Text(label), subtitle: Text(value)),
    );
  }
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
