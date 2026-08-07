import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:r3_ti_faceattend/src/core/network/authenticated_api_client.dart';
import 'package:r3_ti_faceattend/src/face/data/face_api_client.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_failure.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_model_config.dart';
import 'package:r3_ti_faceattend/src/face/domain/face_status.dart';

void main() {
  test('getStatus parses NOT_ENROLLED', () async {
    final api = HttpFaceApiClient(
      client: FakeRequester(
        response: response(<String, Object?>{
          'enrolled': false,
          'face_status': 'NOT_ENROLLED',
        }),
      ),
    );

    final status = await api.getStatus();

    expect(status.status, FaceEnrollmentStatus.notEnrolled);
  });

  test('getStatus parses ENROLLED', () async {
    final api = HttpFaceApiClient(
      client: FakeRequester(
        response: response(<String, Object?>{
          'enrolled': true,
          'face_status': 'ENROLLED',
          'embedding_model': FaceModelConfig.identifier,
          'embedding_version': FaceModelConfig.version,
          'enrolled_at': '2026-08-07T01:00:00Z',
        }),
      ),
    );

    final status = await api.getStatus();

    expect(status.status, FaceEnrollmentStatus.enrolled);
    expect(status.enrolledAt, isNotNull);
  });

  test('enroll sends only embedding and model metadata', () async {
    final requester = FakeRequester(
      response: response(<String, Object?>{
        'enrolled': true,
        'face_status': 'ENROLLED',
      }),
    );
    final api = HttpFaceApiClient(client: requester);

    await api.enroll(
      embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 0.1),
      embeddingModel: FaceModelConfig.identifier,
      embeddingVersion: FaceModelConfig.version,
    );

    expect(requester.lastMethod, 'POST');
    expect(requester.lastPath, '/face/enroll');
    expect(
      requester.lastBody?.keys,
      containsAll(['embedding', 'embedding_model', 'embedding_version']),
    );
    expect(requester.lastBody?.containsKey('user_id'), isFalse);
    expect(requester.lastBody?.containsKey('email'), isFalse);
  });

  test('verify sends model metadata and parses verified result', () async {
    final requester = FakeRequester(
      response: response(<String, Object?>{'verified': true}),
    );
    final api = HttpFaceApiClient(client: requester);

    final result = await api.verify(
      embedding: List<double>.filled(FaceModelConfig.embeddingDimension, 0.1),
      embeddingModel: FaceModelConfig.identifier,
      embeddingVersion: FaceModelConfig.version,
    );

    expect(result.verified, isTrue);
    expect(requester.lastMethod, 'POST');
    expect(requester.lastPath, '/face/verify');
    expect(
      requester.lastBody?.keys,
      containsAll(['embedding', 'embedding_model', 'embedding_version']),
    );
    expect(requester.lastBody?.containsKey('user_id'), isFalse);
    expect(requester.lastBody?.containsKey('threshold'), isFalse);
    expect(requester.lastBody?.containsKey('verified'), isFalse);
  });

  test('verify maps not enrolled conflict and malformed response', () async {
    final notEnrolled = HttpFaceApiClient(
      client: FakeRequester(
        response: AuthenticatedApiResponse(
          statusCode: HttpStatus.conflict,
          payload: <String, Object?>{
            'status': 'error',
            'message': 'wajah belum terdaftar',
          },
        ),
      ),
    );
    await expectLater(
      notEnrolled.verify(
        embedding: const [0.1],
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      ),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.notEnrolled,
        ),
      ),
    );

    final malformed = HttpFaceApiClient(
      client: FakeRequester(
        response: const AuthenticatedApiResponse(
          statusCode: HttpStatus.ok,
          payload: <String, Object?>{
            'status': 'ok',
            'data': <String, Object?>{},
          },
        ),
      ),
    );
    await expectLater(
      malformed.verify(
        embedding: const [0.1],
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      ),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.malformedResponse,
        ),
      ),
    );
  });

  test('maps duplicate, offline and malformed response', () async {
    final duplicate = HttpFaceApiClient(
      client: FakeRequester(
        response: AuthenticatedApiResponse(
          statusCode: HttpStatus.conflict,
          payload: <String, Object?>{
            'status': 'error',
            'message': 'wajah sudah terdaftar',
          },
        ),
      ),
    );
    await expectLater(
      duplicate.enroll(
        embedding: const [0.1],
        embeddingModel: FaceModelConfig.identifier,
        embeddingVersion: FaceModelConfig.version,
      ),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.duplicateEnrollment,
        ),
      ),
    );

    final offline = HttpFaceApiClient(
      client: FakeRequester(
        failure: const AuthenticatedApiFailure(
          AuthenticatedApiFailureKind.apiUnavailable,
          'offline',
        ),
      ),
    );
    await expectLater(
      offline.getStatus(),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.apiUnavailable,
        ),
      ),
    );

    final malformed = HttpFaceApiClient(
      client: FakeRequester(
        response: const AuthenticatedApiResponse(
          statusCode: HttpStatus.ok,
          payload: <String, Object?>{
            'status': 'ok',
            'data': <String, Object?>{},
          },
        ),
      ),
    );
    await expectLater(
      malformed.getStatus(),
      throwsA(
        isA<FaceFailure>().having(
          (e) => e.kind,
          'kind',
          FaceFailureKind.malformedResponse,
        ),
      ),
    );
  });
}

AuthenticatedApiResponse response(Map<String, Object?> data) {
  return AuthenticatedApiResponse(
    statusCode: HttpStatus.ok,
    payload: <String, Object?>{'status': 'ok', 'data': data},
  );
}

class FakeRequester implements AuthenticatedRequester {
  FakeRequester({this.response, this.failure});

  final AuthenticatedApiResponse? response;
  final AuthenticatedApiFailure? failure;
  String? lastMethod;
  String? lastPath;
  Map<String, Object?>? lastBody;

  @override
  Future<AuthenticatedApiResponse> send({
    required String method,
    required String path,
    Map<String, String>? queryParameters,
    Map<String, Object?>? body,
  }) async {
    lastMethod = method;
    lastPath = path;
    lastBody = body;
    final thrown = failure;
    if (thrown != null) {
      throw thrown;
    }
    return response!;
  }
}
