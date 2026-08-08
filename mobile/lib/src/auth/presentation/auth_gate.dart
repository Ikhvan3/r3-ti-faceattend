import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'auth_controller.dart';
import 'home_page.dart';
import 'login_page.dart';

class AuthGate extends StatelessWidget {
  const AuthGate({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthController>(
      builder: (context, controller, _) {
        switch (controller.status) {
          case AuthControllerStatus.initial:
          case AuthControllerStatus.loading:
            return const SplashPage();
          case AuthControllerStatus.authenticated:
            final user = controller.user;
            if (user == null || !user.isUser || !user.isActive) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                controller.logout();
              });
              return const SplashPage();
            }
            // The key forces a fresh Home subtree (including attendance and
            // face controllers) whenever the authenticated user changes. This
            // prevents enrollment/status state from one account being reused
            // by another account during rapid logout/login testing.
            return HomePage(key: ValueKey<String>(user.id), user: user);
          case AuthControllerStatus.unauthenticated:
          case AuthControllerStatus.failure:
            if (controller.sessionRestoreFailed) {
              return SessionRecoveryPage(
                message:
                    controller.errorMessage ??
                    'Session tersimpan, tetapi backend belum dapat dihubungi.',
              );
            }
            return const LoginPage();
        }
      },
    );
  }
}

class SessionRecoveryPage extends StatelessWidget {
  const SessionRecoveryPage({required this.message, super.key});

  final String message;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Icon(
                    Icons.wifi_off_rounded,
                    size: 48,
                    color: theme.colorScheme.primary,
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Belum terhubung ke backend',
                    textAlign: TextAlign.center,
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    message,
                    textAlign: TextAlign.center,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 24),
                  FilledButton.icon(
                    onPressed: () => context.read<AuthController>().initialize(),
                    icon: const Icon(Icons.refresh_rounded),
                    label: const Text('Coba lagi'),
                  ),
                  const SizedBox(height: 8),
                  TextButton(
                    onPressed: () => context.read<AuthController>().clearError(),
                    child: const Text('Masuk dengan akun lain'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class SplashPage extends StatelessWidget {
  const SplashPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 16),
            Text('Memeriksa session...'),
          ],
        ),
      ),
    );
  }
}
