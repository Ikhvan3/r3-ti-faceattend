import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:provider/provider.dart';

import 'src/auth/data/auth_api_client.dart';
import 'src/auth/data/auth_repository.dart';
import 'src/auth/data/token_storage.dart';
import 'src/auth/presentation/auth_controller.dart';
import 'src/auth/presentation/auth_gate.dart';
import 'src/config/api_config.dart';

void main() {
  final apiConfig = ApiConfig.fromEnvironment();
  final tokenStorage = SecureTokenStorage();
  final apiClient = HttpAuthApiClient(
    baseUrl: apiConfig.baseUrl,
    client: http.Client(),
  );
  final repository = AuthRepository(api: apiClient, tokenStorage: tokenStorage);
  if (kDebugMode) {
    debugPrint('API base URL: ${apiConfig.baseUrl}');
  }

  runApp(R3TiFaceAttendApp(repository: repository));
}

class R3TiFaceAttendApp extends StatelessWidget {
  const R3TiFaceAttendApp({required this.repository, super.key});

  final AuthRepository repository;

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider<AuthController>(
      create: (_) => AuthController(repository)..initialize(),
      child: MaterialApp(
        title: 'R3 TI FaceAttend',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(seedColor: Colors.green),
          inputDecorationTheme: const InputDecorationTheme(
            border: OutlineInputBorder(),
          ),
          useMaterial3: true,
        ),
        home: const AuthGate(),
      ),
    );
  }
}
