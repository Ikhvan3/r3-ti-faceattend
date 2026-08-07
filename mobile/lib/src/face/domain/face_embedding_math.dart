import 'dart:math';

import 'face_failure.dart';

List<double> l2NormalizeEmbedding(List<double> embedding) {
  if (embedding.isEmpty || embedding.any((value) => !value.isFinite)) {
    throw const FaceFailure(
      FaceFailureKind.invalidEmbedding,
      'Embedding wajah tidak valid.',
    );
  }
  final norm = sqrt(
    embedding.fold<double>(0, (sum, value) => sum + value * value),
  );
  if (!norm.isFinite || norm == 0) {
    throw const FaceFailure(
      FaceFailureKind.invalidEmbedding,
      'Embedding wajah tidak valid.',
    );
  }
  return embedding.map((value) => value / norm).toList(growable: false);
}

List<double> averageEmbeddings(
  List<List<double>> samples, {
  required int dimension,
}) {
  if (samples.isEmpty) {
    throw const FaceFailure(
      FaceFailureKind.invalidEmbedding,
      'Sample wajah belum tersedia.',
    );
  }
  final totals = List<double>.filled(dimension, 0);
  for (final sample in samples) {
    if (sample.length != dimension) {
      throw const FaceFailure(
        FaceFailureKind.invalidEmbedding,
        'Dimensi embedding wajah tidak valid.',
      );
    }
    for (var i = 0; i < dimension; i += 1) {
      totals[i] += sample[i];
    }
  }
  return l2NormalizeEmbedding(
    totals.map((value) => value / samples.length).toList(growable: false),
  );
}
