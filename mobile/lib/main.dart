import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:provider/provider.dart';

import 'src/auth/data/auth_api_client.dart';
import 'src/auth/data/auth_repository.dart';
import 'src/auth/data/token_storage.dart';
import 'src/auth/presentation/auth_controller.dart';
import 'src/auth/presentation/auth_gate.dart';
import 'src/attendance/data/attendance_api_client.dart';
import 'src/attendance/data/attendance_repository.dart';
import 'src/attendance/data/location_service.dart';
import 'src/core/network/authenticated_api_client.dart';
import 'src/config/api_config.dart';
import 'src/face/data/face_api_client.dart';
import 'src/face/data/face_repository.dart';
import 'src/face/domain/face_attendance_grant.dart';
import 'src/face/domain/face_failure.dart';
import 'src/face/domain/face_status.dart';
import 'src/face/domain/face_verification_result.dart';

void main() {
  final apiConfig = ApiConfig.fromEnvironment();
  final tokenStorage = SecureTokenStorage();
  final httpClient = http.Client();
  final apiClient = HttpAuthApiClient(
    baseUrl: apiConfig.baseUrl,
    client: httpClient,
  );
  final authenticatedClient = AuthenticatedApiClient(
    baseUrl: apiConfig.baseUrl,
    client: httpClient,
    tokenStorage: tokenStorage,
    authApi: apiClient,
  );
  final authRepository = AuthRepository(
    api: apiClient,
    tokenStorage: tokenStorage,
  );
  final attendanceRepository = AttendanceRepository(
    api: HttpAttendanceApiClient(client: authenticatedClient),
  );
  final faceRepository = FaceRepository(
    api: HttpFaceApiClient(client: authenticatedClient),
  );
  if (kDebugMode) {
    debugPrint('API base URL: ${apiConfig.baseUrl}');
  }

  runApp(
    R3TiFaceAttendApp(
      authRepository: authRepository,
      attendanceRepository: attendanceRepository,
      faceRepository: faceRepository,
    ),
  );
}

class R3TiFaceAttendApp extends StatelessWidget {
  const R3TiFaceAttendApp({
    required this.authRepository,
    required this.attendanceRepository,
    this.faceRepository = const FaceRepository(api: _UnavailableFaceApi()),
    this.locationService = const GeolocatorLocationService(),
    super.key,
  });

  final AuthRepository authRepository;
  final AttendanceRepository attendanceRepository;
  final FaceRepository faceRepository;
  final LocationService locationService;

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        Provider<AttendanceRepository>.value(value: attendanceRepository),
        Provider<FaceRepository>.value(value: faceRepository),
        Provider<LocationService>.value(value: locationService),
        ChangeNotifierProvider<AuthController>(
          create: (_) => AuthController(authRepository)..initialize(),
        ),
      ],
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

class _UnavailableFaceApi implements FaceApi {
  const _UnavailableFaceApi();

  @override
  Future<FaceStatus> getStatus() {
    throw const FaceFailure(
      FaceFailureKind.apiUnavailable,
      'Status wajah belum tersedia.',
    );
  }

  @override
  Future<FaceStatus> enroll({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    throw const FaceFailure(
      FaceFailureKind.apiUnavailable,
      'Enrollment wajah belum tersedia.',
    );
  }

  @override
  Future<void> resetEnrollment() {
    throw const FaceFailure(
      FaceFailureKind.apiUnavailable,
      'Reset wajah belum tersedia.',
    );
  }

  @override
  Future<FaceVerificationResult> verify({
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    throw const FaceFailure(
      FaceFailureKind.apiUnavailable,
      'Verifikasi wajah belum tersedia.',
    );
  }

  @override
  Future<FaceAttendanceGrant> verifyForAttendance({
    required String purpose,
    required List<double> embedding,
    required String embeddingModel,
    required String embeddingVersion,
  }) {
    throw const FaceFailure(
      FaceFailureKind.apiUnavailable,
      'Verifikasi wajah belum tersedia.',
    );
  }
}
