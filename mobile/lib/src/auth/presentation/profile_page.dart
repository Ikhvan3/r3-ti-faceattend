import 'package:flutter/material.dart';

import '../domain/user_profile.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({required this.user, super.key});

  final UserProfile user;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Profil Pegawai')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          _ProfileItem(label: 'Nama', value: user.name),
          _ProfileItem(label: 'Nomor Pegawai', value: user.employeeNumber),
          _ProfileItem(label: 'Email', value: user.email),
          _ProfileItem(label: 'Telepon', value: user.phone ?? '-'),
          _ProfileItem(label: 'Jabatan', value: user.position ?? '-'),
          _ProfileItem(label: 'Role', value: user.role),
          _ProfileItem(label: 'Status Akun', value: user.accountStatus),
        ],
      ),
    );
  }
}

class _ProfileItem extends StatelessWidget {
  const _ProfileItem({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(title: Text(label), subtitle: Text(value)),
    );
  }
}
