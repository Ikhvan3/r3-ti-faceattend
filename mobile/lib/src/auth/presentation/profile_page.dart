import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../domain/user_profile.dart';
import 'auth_controller.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({required this.user, this.embedded = false, super.key});

  final UserProfile user;
  final bool embedded;

  @override
  Widget build(BuildContext context) {
    if (embedded) {
      return CustomScrollView(
        slivers: [
          const SliverAppBar(pinned: true, title: Text('Profil')),
          SliverPadding(
            padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
            sliver: SliverList.list(children: _content(context)),
          ),
        ],
      );
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Profil Pegawai')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: _content(context),
      ),
    );
  }

  List<Widget> _content(BuildContext context) {
    final theme = Theme.of(context);

    return [
      Card(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Row(
            children: [
              CircleAvatar(
                radius: 28,
                backgroundColor: theme.colorScheme.primary,
                foregroundColor: theme.colorScheme.onPrimary,
                child: Text(_initials(user.name)),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      user.name,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      user.position ?? 'Pegawai TI',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
      const SizedBox(height: 14),
      _ProfileItem(
        icon: Icons.badge_outlined,
        label: 'Nomor Pegawai',
        value: user.employeeNumber,
      ),
      _ProfileItem(
        icon: Icons.mail_outline_rounded,
        label: 'Email',
        value: user.email,
      ),
      _ProfileItem(
        icon: Icons.phone_outlined,
        label: 'Telepon',
        value: user.phone ?? '-',
      ),
      _ProfileItem(
        icon: Icons.verified_user_outlined,
        label: 'Status Akun',
        value: _statusLabel(user.accountStatus),
      ),
      const SizedBox(height: 18),
      OutlinedButton.icon(
        onPressed: context.watch<AuthController>().isSubmitting
            ? null
            : () => _confirmLogout(context),
        icon: const Icon(Icons.logout_rounded),
        label: const Text('Logout'),
      ),
    ];
  }

  Future<void> _confirmLogout(BuildContext context) async {
    final confirmed = await showModalBottomSheet<bool>(
      context: context,
      showDragHandle: true,
      builder: (sheetContext) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'Keluar dari aplikasi?',
                  style: Theme.of(sheetContext).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Session perangkat akan dihapus. Anda dapat masuk kembali dengan email dan password.',
                ),
                const SizedBox(height: 18),
                FilledButton.icon(
                  onPressed: () => Navigator.of(sheetContext).pop(true),
                  icon: const Icon(Icons.logout_rounded),
                  label: const Text('Logout'),
                ),
                const SizedBox(height: 8),
                TextButton(
                  onPressed: () => Navigator.of(sheetContext).pop(false),
                  child: const Text('Batal'),
                ),
              ],
            ),
          ),
        );
      },
    );

    if (confirmed == true && context.mounted) {
      await context.read<AuthController>().logout();
    }
  }
}

class _ProfileItem extends StatelessWidget {
  const _ProfileItem({
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

    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Card(
        child: ListTile(
          leading: Icon(icon, color: theme.colorScheme.primary),
          title: Text(label),
          subtitle: Text(value),
        ),
      ),
    );
  }
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
