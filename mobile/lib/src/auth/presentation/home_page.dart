import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

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
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            'Halo, ${user.name}',
            style: Theme.of(
              context,
            ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 8),
          Text('Akun ${_statusLabel(user.accountStatus)}'),
          const SizedBox(height: 20),
          _InfoCard(label: 'Nomor Pegawai', value: user.employeeNumber),
          _InfoCard(label: 'Jabatan', value: user.position ?? '-'),
          _InfoCard(label: 'Email', value: user.email),
          const SizedBox(height: 20),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Modul Absensi',
                    style: TextStyle(fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Check-in, check-out, lokasi, dan verifikasi wajah akan dibuat pada tahap berikutnya.',
                  ),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: null,
                    child: const Text('Check-in belum tersedia'),
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
