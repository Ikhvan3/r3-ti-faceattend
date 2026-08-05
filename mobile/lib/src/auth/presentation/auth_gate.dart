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
            return HomePage(user: user);
          case AuthControllerStatus.unauthenticated:
          case AuthControllerStatus.failure:
            return const LoginPage();
        }
      },
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
